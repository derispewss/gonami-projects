-- +goose Up
-- +goose StatementBegin

INSERT INTO categories (name, type) VALUES
    ('Food & Beverage', 'expense'),
    ('Transportation',  'expense'),
    ('Shopping',        'expense'),
    ('Bills',           'expense'),
    ('Entertainment',   'expense'),
    ('Health',          'expense'),
    ('Education',       'expense'),
    ('Travel',          'expense'),
    ('Subscription',    'expense'),
    ('Other',           'expense');

INSERT INTO categories (name, type) VALUES
    ('Salary',          'income'),
    ('Freelance',       'income'),
    ('Business',        'income'),
    ('Gift',            'income'),
    ('Investment',      'income'),
    ('Other',           'income');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM categories WHERE user_id IS NULL;

-- +goose StatementEnd
