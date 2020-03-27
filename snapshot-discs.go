package main

import (
	"context"
	"encoding/json"
	"fmt"
	_ "fmt"
	"os"
	"strconv"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk"
)

func SnapshotHandler(ctx context.Context, event MessageQueueEvent) error {
	// Авторизация в SDK при помощи сервисного аккаунта
	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		// Вызов InstanceServiceAccount автоматически запрашивает IAM-токен и формирует
		// при помощи него данные для авторизации в SDK
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return err
	}
	now := time.Now()
	// Получаем значение периода жизни снепшота из переменной окружения
	ttl, err := strconv.Atoi(os.Getenv("TTL"))
	if err != nil {
		return err
	}

	// Вычисляем таймстемп, после которого можно будет удалить снепшот.
	expirationTs := strconv.Itoa(int(now.Unix()) + ttl)

	// Парсим json с данными в каком фолдере какой диск нужно снепшотить
	body := event.Messages[0].Details.Message.Body
	createSnapshotParams := &CreateSnapshotParams{}
	err = json.Unmarshal([]byte(body), createSnapshotParams)
	if err != nil {
		return err
	}

	// При помощи YC SDK создаем снепшот, указывая в лейблах время его жизни.
	// Он не будет удален автоматически Облаком. Этим будет заниматься функция описанная в ./delete-expired.go
	snapshotOp, err := sdk.Compute().Snapshot().Create(ctx, &compute.CreateSnapshotRequest{
		FolderId: createSnapshotParams.FolderId,
		DiskId:   createSnapshotParams.DiskId,
		Labels: map[string]string{
			"expiration_ts": expirationTs,
		},
	})
	if err != nil {
		return err
	}
	// Если снепшот по каким-то причинам создать не удалось, сообщение вернется в очередь. После этого триггер
	// снова возьмет его из очереди, вызовет эту функцию снова и будет сделана еще одна попытка его создать.
	if opErr := snapshotOp.GetError(); opErr != nil {
		return fmt.Errorf(opErr.Message)
	}

	return nil
}
