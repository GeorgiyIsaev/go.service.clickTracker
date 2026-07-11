CREATE TABLE raw_clicks (
                            id BIGSERIAL PRIMARY KEY,
                            author_id VARCHAR(255) NOT NULL,
                            user_id VARCHAR(255) NOT NULL,
                            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                            UNIQUE(author_id, user_id, created_at::date)  -- один пользователь может кликнуть только раз в сутки
);

CREATE TABLE daily_stats (
                             date DATE NOT NULL,
                             author_id VARCHAR(255) NOT NULL,
                             unique_users INT NOT NULL,
                             PRIMARY KEY (date, author_id)
);