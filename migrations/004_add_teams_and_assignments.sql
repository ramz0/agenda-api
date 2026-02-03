-- +migrate Up

-- 1. Update user roles: convert speaker and attendee to user
ALTER TABLE users ALTER COLUMN role DROP DEFAULT;
ALTER TABLE users ALTER COLUMN role TYPE text;
UPDATE users SET role = 'user' WHERE role IN ('speaker', 'attendee');
DROP TYPE IF EXISTS user_role;
CREATE TYPE user_role AS ENUM ('admin', 'user');
ALTER TABLE users ALTER COLUMN role TYPE user_role USING role::user_role;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'user';

-- 2. Create event type enum
CREATE TYPE event_type AS ENUM ('personal', 'team');

-- 3. Create teams table
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_teams_created_by ON teams(created_by);

-- 4. Create team_members table
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, user_id)
);

CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);

-- 5. Add type and team_id columns to events
ALTER TABLE events ADD COLUMN type event_type NOT NULL DEFAULT 'personal';
ALTER TABLE events ADD COLUMN team_id UUID REFERENCES teams(id) ON DELETE SET NULL;

-- Make capacity optional (null = no limit)
ALTER TABLE events ALTER COLUMN capacity DROP NOT NULL;
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_capacity_check;

-- Remove speaker_id (no longer needed)
ALTER TABLE events DROP COLUMN IF EXISTS speaker_id;

CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_team_id ON events(team_id);

-- 6. Create assignment status enum
CREATE TYPE assignment_status AS ENUM ('pending', 'approved', 'rejected');

-- 7. Create event_assignments table
CREATE TABLE event_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status assignment_status NOT NULL DEFAULT 'pending',
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(event_id, user_id)
);

CREATE INDEX idx_event_assignments_event_id ON event_assignments(event_id);
CREATE INDEX idx_event_assignments_user_id ON event_assignments(user_id);
CREATE INDEX idx_event_assignments_status ON event_assignments(status);

-- +migrate Down
DROP TABLE IF EXISTS event_assignments;
DROP TYPE IF EXISTS assignment_status;

ALTER TABLE events DROP COLUMN IF EXISTS team_id;
ALTER TABLE events DROP COLUMN IF EXISTS type;
DROP TYPE IF EXISTS event_type;

DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;

-- Restore old role enum
ALTER TABLE users ALTER COLUMN role DROP DEFAULT;
ALTER TABLE users ALTER COLUMN role TYPE text;
DROP TYPE IF EXISTS user_role;
CREATE TYPE user_role AS ENUM ('admin', 'speaker', 'attendee');
ALTER TABLE users ALTER COLUMN role TYPE user_role USING 'attendee'::user_role;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'attendee';
