
CREATE TABLE IF NOT EXISTS users_data (
                                          user_id       UUID PRIMARY KEY,
                                          user_name     TEXT        NOT NULL UNIQUE,
                                          user_password TEXT        NOT NULL,
                                          user_role     TEXT        NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
                                      order_id     UUID PRIMARY KEY,
                                      user_id      UUID        NOT NULL,
                                      market_id    UUID        NOT NULL,
                                      price        NUMERIC      NOT NULL,
                                      amount       NUMERIC      NOT NULL,
                                      order_status TEXT         NOT NULL,
                                      created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS idx_orders_user_id_order_id ON orders (user_id, order_id);

CREATE TABLE IF NOT EXISTS markets (
                                       market_id            UUID PRIMARY KEY,
                                       market_name          TEXT        NOT NULL UNIQUE,
                                       goods_id             UUID        NOT NULL,
                                       accessibility        BOOLEAN     NOT NULL DEFAULT TRUE,
                                       ttl                  TIMESTAMPTZ                   
);
