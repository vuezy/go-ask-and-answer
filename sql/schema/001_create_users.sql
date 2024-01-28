-- +goose Up
CREATE TABLE users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  `name` VARCHAR(20) NOT NULL,
  email VARCHAR(50) UNIQUE NOT NULL,
  `password` VARCHAR(256) NOT NULL,
  points INT NOT NULL DEFAULT 0,
  credits INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

-- +goose Down
DROP TABLE users;