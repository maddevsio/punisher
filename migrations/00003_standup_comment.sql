-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup` MODIFY COLUMN `comment` TEXT;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standup` MODIFY COLUMN `comment` VARCHAR(255) NOT NULL;