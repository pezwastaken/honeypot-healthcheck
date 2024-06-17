MAIN_PATH := ./cmd/
BINARY_NAME := healthcheck
USER := user
HOUR_STEP := 6
BINARY_PATH := /home/${USER}/honeypot-healthcheck/bin
CRON_ENTRY := 0 */${HOUR_STEP} * * * ${BINARY_PATH}/${BINARY_NAME}

build:
	go build -o bin/${BINARY_NAME} ${MAIN_PATH}

addcron:
	echo "${CRON_ENTRY}" | crontab -

all: build addcron
