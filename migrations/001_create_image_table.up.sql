CREATE TYPE status_type AS ENUM ('failed', 'processing', 'processed', 'new');
CREATE TYPE mime_type AS ENUM ('image/png', 'image/jpg', 'image/gif');
CREATE TYPE process_type AS ENUM('watermark', 'resize', 'miniature');

CREATE TABLE IF NOT EXISTS image (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename TEXT NOT NULL ,
    height INT NOT NULL,
    width  INT NOT NULL,
    size BIGINT NOT NULL,

    process_type process_type NOT NULL ,
    mime_type mime_type NOT NULL ,
    status status_type NOT NULL DEFAULT 'new',

    updated_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    processed_at timestamptz
);
