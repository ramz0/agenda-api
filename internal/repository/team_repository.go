package repository

import (
	"agenda-api/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(team *models.Team) error {
	query := `
		INSERT INTO teams (id, name, description, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowx(
		query,
		team.ID, team.Name, team.Description, team.CreatedBy,
		team.CreatedAt, team.UpdatedAt,
	).Scan(&team.ID, &team.CreatedAt, &team.UpdatedAt)
}

func (r *TeamRepository) GetByID(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	query := `SELECT * FROM teams WHERE id = $1`
	err := r.db.Get(&team, query, id)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) GetAll() ([]models.Team, error) {
	var teams []models.Team
	query := `SELECT * FROM teams ORDER BY name`
	err := r.db.Select(&teams, query)
	return teams, err
}

func (r *TeamRepository) GetByCreatedBy(userID uuid.UUID) ([]models.Team, error) {
	var teams []models.Team
	query := `SELECT * FROM teams WHERE created_by = $1 ORDER BY name`
	err := r.db.Select(&teams, query, userID)
	return teams, err
}

func (r *TeamRepository) GetByMemberUserID(userID uuid.UUID) ([]models.Team, error) {
	var teams []models.Team
	query := `
		SELECT t.* FROM teams t
		INNER JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1
		ORDER BY t.name`
	err := r.db.Select(&teams, query, userID)
	return teams, err
}

func (r *TeamRepository) Update(team *models.Team) error {
	query := `
		UPDATE teams
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4`

	team.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, team.Name, team.Description, team.UpdatedAt, team.ID)
	return err
}

func (r *TeamRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *TeamRepository) AddMember(member *models.TeamMember) error {
	query := `
		INSERT INTO team_members (id, team_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return r.db.QueryRowx(
		query,
		member.ID, member.TeamID, member.UserID, member.CreatedAt,
	).Scan(&member.ID, &member.CreatedAt)
}

func (r *TeamRepository) RemoveMember(teamID, userID uuid.UUID) error {
	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`
	_, err := r.db.Exec(query, teamID, userID)
	return err
}

func (r *TeamRepository) GetMembers(teamID uuid.UUID) ([]models.TeamMemberWithUser, error) {
	var members []models.TeamMemberWithUser
	query := `
		SELECT tm.*, u.name as user_name, u.email as user_email
		FROM team_members tm
		INNER JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1
		ORDER BY u.name`
	err := r.db.Select(&members, query, teamID)
	return members, err
}

func (r *TeamRepository) IsMember(teamID, userID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM team_members WHERE team_id = $1 AND user_id = $2`
	err := r.db.Get(&count, query, teamID, userID)
	return count > 0, err
}
