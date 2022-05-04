#!/bin/sh

if [[ -f .env ]]
then
    set -a
    . .env
    set +a
fi

GO111MODULE=on go run ./scripts/queue-create.go

yc serverless function create --name=spawn-snapshot-tasks
yc serverless function create --name=snapshot-disks
yc serverless function create --name=delete-expired-snapshots

yc serverless trigger create timer \
    --name=spawn-snapshot-tasks \
    --cron-expression="$CREATE_CRON" \
    --invoke-function-name=spawn-snapshot-tasks \
    --invoke-function-tag="\$latest" \
    --invoke-function-service-account-id=$SERVICE_ACCOUNT_ID

yc serverless trigger create message-queue \
    --name=snapshot-disks \
    --queue=$QUEUE_ARN \
    --queue-service-account-id $SERVICE_ACCOUNT_ID \
    --invoke-function-name=snapshot-disks \
    --invoke-function-tag="\$latest" \
    --invoke-function-service-account-id=$SERVICE_ACCOUNT_ID

yc serverless trigger create timer \
    --name=delete-expired-snapshots \
    --cron-expression="$DELETE_CRON" \
    --invoke-function-name=delete-expired-snapshots \
    --invoke-function-tag="\$latest" \
    --invoke-function-service-account-id=$SERVICE_ACCOUNT_ID

