-- +goose Up
-- +goose StatementBegin
CREATE TABLE providers
(
    id         BINARY(16)   NOT NULL,
    name       varchar(100) NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NULL     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP    NULL     DEFAULT NULL,
    PRIMARY KEY (id)
)
    ENGINE = 'InnoDB'
    DEFAULT CHARSET = 'utf8mb4'
    COLLATE = 'utf8mb4_unicode_ci';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE overflow_providers
(
    id          BINARY(16)   NOT NULL,
    provider_id BINARY(16)   NOT NULL,
    fallback_id varchar(100) NOT NULL,
    increment   INT                   DEFAULT 1,
    name        varchar(100) NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP    NULL     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP    NULL     DEFAULT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (provider_id) REFERENCES providers (id)
)
    ENGINE = 'InnoDB'
    DEFAULT CHARSET = 'utf8mb4'
    COLLATE = 'utf8mb4_unicode_ci';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE subscribers
(
    id                   BINARY(16)   NOT NULL,
    overflow_provider_id BINARY(16)   NOT NULL,
    name                 varchar(100) NOT NULL,
    email                varchar(100) NOT NULL,
    created_at           TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP    NULL     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at           TIMESTAMP    NULL     DEFAULT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (overflow_provider_id) REFERENCES overflow_providers (id)
)
    ENGINE = 'InnoDB'
    DEFAULT CHARSET = 'utf8mb4'
    COLLATE = 'utf8mb4_unicode_ci';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE api_keys
(
    id            BINARY(16)   NOT NULL,
    subscriber_id BINARY(16)   NOT NULL,
    api_key       varchar(100) NOT NULL UNIQUE,
    admin         bool         NOT NULL,
    created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP    NULL     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP    NULL     DEFAULT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (subscriber_id) REFERENCES subscribers (id)
)
    ENGINE = 'InnoDB'
    DEFAULT CHARSET = 'utf8mb4'
    COLLATE = 'utf8mb4_unicode_ci';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE tanks
(
    id             BINARY(16)      NOT NULL,
    subscriber_id  BINARY(16)      NOT NULL,
    name           varchar(100)    NOT NULL,
    enabled        bool                     DEFAULT FALSE,
    max_tank_size  INT             NOT NULL,
    default_number BIGINT UNSIGNED NOT NULL,
    default_ttl    INT                      DEFAULT 600,
    fallback_url   varchar(256),
    created_at     TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP       NULL     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at     TIMESTAMP       NULL     DEFAULT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (subscriber_id) REFERENCES subscribers (id)
)
    ENGINE = 'InnoDB'
    DEFAULT CHARSET = 'utf8mb4'
    COLLATE = 'utf8mb4_unicode_ci';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE phone_numbers
(
    id               BINARY(16)      NOT NULL,
    provider_id      BINARY(16)      NOT NULL,
    tank_id          BINARY(16)      NOT NULL,
    phone_number     BIGINT UNSIGNED NOT NULL UNIQUE,
    last_reservation TIMESTAMP       NULL     DEFAULT NULL,
    last_call        TIMESTAMP       NULL     DEFAULT NULL,
    created_at       TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP       NULL     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at       TIMESTAMP       NULL     DEFAULT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (provider_id) REFERENCES providers (id),
    FOREIGN KEY (tank_id) REFERENCES tanks (id)
)
    ENGINE = 'InnoDB'
    DEFAULT CHARSET = 'utf8mb4'
    COLLATE = 'utf8mb4_unicode_ci';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE reservations
(
    id               BINARY(16)      NOT NULL,
    phone_number_id  BINARY(16)      NOT NULL,
    tank_id          BINARY(16)      NOT NULL,
    ani_phone_number BIGINT UNSIGNED NOT NULL,
    expiration       TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    meta             blob,
    created_at       TIMESTAMP(2)    NOT NULL DEFAULT CURRENT_TIMESTAMP(2),
    updated_at       TIMESTAMP(2)    NULL     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP(2),
    deleted_at       TIMESTAMP(2)    NULL     DEFAULT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (phone_number_id) REFERENCES phone_numbers (id),
    FOREIGN KEY (tank_id) REFERENCES tanks (id)
)
    ENGINE = 'InnoDB'
    DEFAULT CHARSET = 'utf8mb4'
    COLLATE = 'utf8mb4_unicode_ci';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE calls
(
    id              BINARY(16)      NOT NULL,
    phone_number_id BINARY(16)      NOT NULL,
    reservation_id  BINARY(16)      NOT NULL,
    ani             BIGINT UNSIGNED NOT NULL,
    destination     BIGINT UNSIGNED NOT NULL,
    created_at      TIMESTAMP(2)    NOT NULL DEFAULT CURRENT_TIMESTAMP(2),
    updated_at      TIMESTAMP(2)    NULL     DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP(2),
    deleted_at      TIMESTAMP(2)    NULL     DEFAULT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (phone_number_id) REFERENCES phone_numbers (id),
    FOREIGN KEY (reservation_id) REFERENCES reservations (id)
)
    ENGINE = 'InnoDB'
    DEFAULT CHARSET = 'utf8mb4'
    COLLATE = 'utf8mb4_unicode_ci';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET FOREIGN_KEY_CHECKS = 0;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS subscribers;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS providers;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS overflow_providers;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS phone_numbers;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS tanks;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS calls;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS api_keys;
-- +goose StatementEnd
-- +goose StatementBegin
SET FOREIGN_KEY_CHECKS = 1;
-- +goose StatementEnd
