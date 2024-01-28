-- +goose Up
CREATE TABLE questions (
  id INT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(50) NOT NULL,
  body VARCHAR(300) NOT NULL,
  `priority_level` INT NOT NULL DEFAULT 0,
  user_id INT NOT NULL,
  responded_at DATETIME NOT NULL,
  closed BOOLEAN NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- +goose Down
DROP TABLE questions;