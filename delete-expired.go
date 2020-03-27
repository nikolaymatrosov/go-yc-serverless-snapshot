package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func DeleteHandler(ctx context.Context) error {
	folderId := os.Getenv("FOLDER_ID")
	// Авторизация в SDK при помощи сервисного аккаунта
	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		// Вызов InstanceServiceAccount автоматически запрашивает IAM-токен и формирует
		// при помощи него данные для авторизации в SDK
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return err
	}

	// Получаем итератор снепшотов при помощи YC SDK
	snapshotIter := sdk.Compute().Snapshot().SnapshotIterator(ctx, folderId)

	// Итрерируемся по нему
	for snapshotIter.Next() {
		snapshot := snapshotIter.Value()
		labels := snapshot.Labels
		if labels == nil {
			continue
		}
		// Проверяем есть ли у снепшота лейбл `expiration_ts`.
		expirationTsVal, ok := labels["expiration_ts"]
		if !ok {
			continue
		}
		now := time.Now()
		expirationTs, err := strconv.Atoi(expirationTsVal)
		if err != nil {
			return nil
		}

		// Если он есть и время сейчас больше, чем то что записано в лейбл, то удаляем снепшот.
		if int(now.Unix()) > expirationTs {
			_, _ = sdk.Compute().Snapshot().Delete(ctx, &compute.DeleteSnapshotRequest{
				SnapshotId: snapshot.Id,
			})
		}
	}

	return nil
}

