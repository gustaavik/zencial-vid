-- Requires PostgreSQL 18+ (native uuidv7() built-in)
-- Switch all primary-key defaults from UUIDv4 to UUIDv7 for better index locality.
ALTER TABLE users              ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE genres             ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE genre_translations ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE videos             ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE plans              ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE user_subscriptions ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE audit_logs         ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE user_sessions      ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE video_cast         ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE casts              ALTER COLUMN id SET DEFAULT uuidv7();
