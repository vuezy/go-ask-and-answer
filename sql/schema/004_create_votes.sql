-- +goose Up
CREATE TABLE votes (
  id INT PRIMARY KEY AUTO_INCREMENT,
  val INT NOT NULL DEFAULT 1,
  answer_id INT NOT NULL,
  user_id INT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY(answer_id) REFERENCES answers(id) ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- +goose Down
DROP TABLE votes;