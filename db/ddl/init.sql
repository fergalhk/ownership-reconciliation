CREATE TABLE resource_owners (
    id           UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    resource_arn TEXT NOT NULL,
    org_group    TEXT NOT NULL,
    created_at   DATE NOT NULL,

    UNIQUE (resource_arn, created_at)
);
