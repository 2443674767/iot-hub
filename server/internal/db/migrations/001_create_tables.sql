-- 设备分类表
CREATE TABLE IF NOT EXISTS device_categories (
    id          BIGSERIAL PRIMARY KEY,
    parent_id   BIGINT REFERENCES device_categories(id),
    model       VARCHAR(64) NOT NULL,
    name        VARCHAR(128) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- 设备表
CREATE TABLE IF NOT EXISTS devices (
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(64) NOT NULL,
    display_name VARCHAR(128),
    can_id       BYTEA NOT NULL UNIQUE,
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    status       VARCHAR(32) NOT NULL DEFAULT 'offline',
    category_id  BIGINT REFERENCES device_categories(id),
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE devices ADD COLUMN IF NOT EXISTS display_name VARCHAR(128);
ALTER TABLE devices ADD COLUMN IF NOT EXISTS enabled BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'offline';
ALTER TABLE devices ADD COLUMN IF NOT EXISTS category_id BIGINT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_devices_category_id'
          AND conrelid = 'devices'::regclass
    ) THEN
        ALTER TABLE devices
            ADD CONSTRAINT fk_devices_category_id
            FOREIGN KEY (category_id)
            REFERENCES device_categories(id);
    END IF;
END $$;

-- 报文日志表
CREATE TABLE IF NOT EXISTS can_messages (
    id          BIGSERIAL PRIMARY KEY,
    can_id      VARCHAR(8) NOT NULL,
    data        VARCHAR(16) NOT NULL,
    direction   SMALLINT NOT NULL,   -- 0: 上行, 1: 下行
    received_at TIMESTAMPTZ DEFAULT NOW()
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'can_messages'
          AND column_name = 'can_id'
          AND data_type = 'bytea'
    ) THEN
        ALTER TABLE can_messages
            ALTER COLUMN can_id TYPE VARCHAR(8)
            USING '0x' || upper(encode(can_id, 'hex'));
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'can_messages'
          AND column_name = 'data'
          AND data_type = 'bytea'
    ) THEN
        ALTER TABLE can_messages
            ALTER COLUMN data TYPE VARCHAR(16)
            USING encode(data, 'hex');
    END IF;
END $$;

-- 原始 CAN 数据表
CREATE TABLE IF NOT EXISTS raw_can_data (
    id          BIGSERIAL PRIMARY KEY,
    can_id      BYTEA,
    direction   SMALLINT,
    read_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    raw_frame   JSONB NOT NULL,
    parsed_data JSONB NOT NULL DEFAULT '{}'::jsonb
);

ALTER TABLE raw_can_data ADD COLUMN IF NOT EXISTS can_id BYTEA;
ALTER TABLE raw_can_data ADD COLUMN IF NOT EXISTS direction SMALLINT;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'raw_can_data'
          AND column_name = 'raw_frame'
          AND data_type = 'bytea'
    ) THEN
        ALTER TABLE raw_can_data
            ALTER COLUMN raw_frame TYPE JSONB
            USING jsonb_build_object(
                'canid',
                CASE
                    WHEN can_id IS NULL OR length(can_id) = 0 THEN ''
                    ELSE '0x' || upper(encode(can_id, 'hex'))
                END,
                'data',
                CASE
                    WHEN length(raw_frame) = 9 THEN encode(substring(raw_frame FROM 2 FOR 8), 'hex')
                    ELSE encode(raw_frame, 'hex')
                END
            );
    END IF;
END $$;

-- TCP 发送目标配置表
CREATE TABLE IF NOT EXISTS tcp_configs (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    host        VARCHAR(128) NOT NULL,
    port        INTEGER NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_tcp_configs_port CHECK (port BETWEEN 1 AND 65535)
);

CREATE INDEX IF NOT EXISTS idx_devices_category_id ON devices(category_id);
CREATE INDEX IF NOT EXISTS idx_device_categories_parent_id ON device_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_can_messages_can_id ON can_messages(can_id);
CREATE INDEX IF NOT EXISTS idx_can_messages_received_at ON can_messages(received_at);
CREATE INDEX IF NOT EXISTS idx_raw_can_data_can_id ON raw_can_data(can_id);
CREATE INDEX IF NOT EXISTS idx_raw_can_data_direction ON raw_can_data(direction);
CREATE INDEX IF NOT EXISTS idx_raw_can_data_read_at ON raw_can_data(read_at);
CREATE INDEX IF NOT EXISTS idx_tcp_configs_enabled ON tcp_configs(enabled);
