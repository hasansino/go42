-- +goose Up
insert or ignore into example_fruits (name) values
('orange'),
('pineapple'),
('watermelon'),
('grapefruit'),
('lime'),
('apple'),
('grapes'),
('pear'),
('banana'),
('peach');

-- +goose Down
truncate example_fruits;
