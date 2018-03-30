-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE `lives` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(255) NOT NULL,
    `lives` INTEGER NOT NULL
);
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE `lives`;