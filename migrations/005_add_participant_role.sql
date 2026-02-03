-- +migrate Up

-- Create participant role enum (speaker, attendee, participant)
-- speaker = ponente (quien presenta)
-- attendee = asistente (quien ayuda al ponente)
-- participant = participante (el público, rol por defecto)
CREATE TYPE participant_role AS ENUM ('speaker', 'attendee', 'participant');

-- Add role column to event_assignments
ALTER TABLE event_assignments ADD COLUMN role participant_role NOT NULL DEFAULT 'participant';

-- +migrate Down
ALTER TABLE event_assignments DROP COLUMN IF EXISTS role;
DROP TYPE IF EXISTS participant_role;
