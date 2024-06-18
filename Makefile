MAIN_PATH := ./cmd/
BINARY_NAME := healthcheck
USER := user
HOUR_STEP := 6
GO_PATH := go
BINARY_PATH := /home/${USER}/honeypot-healthcheck/bin
CRON_ENTRY := 0 */${HOUR_STEP} * * * ${BINARY_PATH}/${BINARY_NAME}

build:
	${GO_PATH} build -o bin/${BINARY_NAME} ${MAIN_PATH}

addcron:
	echo "${CRON_ENTRY}" | crontab -

all: build addcron
