CREATE TABLE raw_clicks (
                            id BIGSERIAL PRIMARY KEY,
                            author_id VARCHAR(255) NOT NULL,
                            user_id VARCHAR(255) NOT NULL,
                            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- Отдельная колонка с датой (без времени) для уникальности по суткам
                            created_date DATE NOT NULL DEFAULT CURRENT_DATE,
    -- Уникальность: один пользователь может кликнуть автора только раз в сутки
                            UNIQUE(author_id, user_id, created_date)
);

CREATE TABLE daily_stats (
                             date DATE NOT NULL,
                             author_id VARCHAR(255) NOT NULL,
                             unique_users INT NOT NULL,
                             PRIMARY KEY (date, author_id)
);