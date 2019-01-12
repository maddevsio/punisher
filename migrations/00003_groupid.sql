-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup` ADD `groupid` INT NOT NULL;
ALTER TABLE `interns` ADD `groupid` INT NOT NULL;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standup` DROP `groupid`;
ALTER TABLE `interns` DROP `groupid`;