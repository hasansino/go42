-- +goose Up
insert into example_fruits (name) values
('orange'),
('pineapple'),
('watermelon'),
('grapefruit'),
('lime'),
('apple'),
('grapes'),
('pear'),
('banana'),
('peach')
on conflict do nothing;

-- +goose Down
truncate example_fruits;
