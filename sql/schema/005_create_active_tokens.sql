-- +goose Up
CREATE TABLE active_tokens (
  sub INT PRIMARY KEY,
  jti VARCHAR(256) NOT NULL,
  UNIQUE(sub, jti),
  FOREIGN KEY(sub) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- +goose Down
DROP TABLE active_tokens;