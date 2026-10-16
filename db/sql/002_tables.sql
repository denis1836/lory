CREATE TABLE IF NOT EXISTS Users (
    usrID BIGSERIAL PRIMARY KEY,
    email VARCHAR(256) NOT NULL UNIQUE,
    password_hash VARCHAR(256) NOT NULL,
    
    name VARCHAR(128),
    description VARCHAR(1024),

    is_verified BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_online TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS User_Sessions (
    usr_sesSID BIGSERIAL PRIMARY KEY,
    usrID BIGINT REFERENCES Users(usrID),

    session_token VARCHAR(256) NOT NULL,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS Projects (
    prjID BIGSERIAL PRIMARY KEY,

    usrID BIGINT REFERENCES Users(usrID), -- owner user ID

    name VARCHAR(256),
    description VARCHAR(1024),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS Tags (
    tgID BIGSERIAL PRIMARY KEY,
    project_ID BIGINT REFERENCES Projects(prjID),

    name VARCHAR(128),
    description VARCHAR(512),

    color CHAR(7) CHECK (color ~ '^#[0-9A-Fa-f]{6}$'), -- #RRGGBB

    is_global BOOL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS Articles (
    artID BIGSERIAL PRIMARY KEY,
    prjID BIGINT REFERENCES Projects(prjID),
    
    slug VARCHAR(256) NOT NULL,
    title VARCHAR(512) NOT NULL,
    
    content TEXT COMPRESSION lz4 NOT NULL,

    created_by BIGINT NOT NULL REFERENCES Users(usrID),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS Article_Revisions (
    artRID BIGSERIAL PRIMARY KEY,
    artID BIGINT REFERENCES Articles(artID) ON DELETE CASCADE,
    usrID BIGINT REFERENCES Users(usrID),
    
    rev_num INTEGER NOT NULL CHECK (rev_num >= 1),
    
    title VARCHAR(512) NOT NULL,
    content TEXT NOT NULL,
    
    content_hash BYTEA NOT NULL CHECK (octet_length(content_hash) = 32),
    message VARCHAR(256),
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT art_revs_uniq UNIQUE (artID, rev_num)
);

CREATE TABLE IF NOT EXISTS Article_Tags (
    artID BIGINT REFERENCES Articles(artID) ON DELETE CASCADE,
    tgID BIGINT REFERENCES Tags(tgID) ON DELETE CASCADE,

    PRIMARY KEY(artID, tgID)
);

CREATE TABLE IF NOT EXISTS Assets (
    astID BIGSERIAL PRIMARY KEY,

    prjID BIGINT REFERENCES Projects(prjID),
    artID BIGINT REFERENCES Articles(artID),
    usrID BIGINT REFERENCES Users(usrID) NOT NULL,

    hash VARCHAR(256),
    asset_type ASSET_TYPE,
    comment VARCHAR(1024),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
