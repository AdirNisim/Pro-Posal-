package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pro-posal/webserver/dao"
	"github.com/pro-posal/webserver/internal/database"
	"github.com/pro-posal/webserver/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm" // Ensure only this version is used
)

type CreateCompanyRequest struct {
	Name       string
	ContactID  string
	Address    string
	LogoBase64 string
}
type UpdateCompanyRequest struct {
	Name       string
	Address    string
	LogoBase64 string
}

type CompanyManagementService interface {
	CreateCompany(context.Context, CreateCompanyRequest) (*models.Company, error)
	DeleteCompany(context.Context, string) (*models.Company, error)
	UpdateCompany(context.Context, string, UpdateCompanyRequest) (*models.Company, error)
	GetCompanies(context.Context, string) ([]*models.Company, error)
}

type CompanyManagementServiceImpl struct {
	db *database.DBConnector
}

func NewCompanyManagementService(db *database.DBConnector) CompanyManagementService {
	return &CompanyManagementServiceImpl{
		db: db,
	}
}

func (s *CompanyManagementServiceImpl) CreateCompany(ctx context.Context, req CreateCompanyRequest) (*models.Company, error) {

	companyDao := dao.Company{
		ID:         uuid.NewString(),
		Name:       req.Name,
		ContactID:  req.ContactID,
		Address:    req.Address,
		LogoBase64: req.LogoBase64,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// TODO: Check if already exists to ensure non-500 errors for idempotency

	err := companyDao.Insert(ctx, s.db.Conn, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("failed inserting user to database: %w", err)
	}

	premmisionDao := dao.Permission{
		ID:         uuid.NewString(),
		UserID:     companyDao.ContactID,
		CompanyID:  companyDao.ID,
		Role:       "Owner",
		ContractID: "AskAlex",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = premmisionDao.Insert(ctx, s.db.Conn, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("failed inserting user premmision, company added: %w", err)
	}

	return companyDaoToCompanyModel(companyDao), nil
}

func (s *CompanyManagementServiceImpl) DeleteCompany(ctx context.Context, id string) (*models.Company, error) {
	companyDao, err := dao.FindCompany(ctx, s.db.Conn, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no company found with ID %s", id)
		}
		return nil, fmt.Errorf("error retrieving company: %w", err)
	}

	company := companyDaoToCompanyModel(*companyDao)

	_, err = companyDao.Delete(ctx, s.db.Conn)
	if err != nil {
		return nil, fmt.Errorf("error deleting company: %w", err)
	}

	return company, nil
}

func (s *CompanyManagementServiceImpl) UpdateCompany(ctx context.Context, id string, req UpdateCompanyRequest) (*models.Company, error) {
	companyDao, err := dao.FindCompany(ctx, s.db.Conn, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no company found with ID %s", id)
		}
		return nil, fmt.Errorf("error retrieving company: %w", err)
	}

	if req.Name != "" {
		companyDao.Name = req.Name
	}
	if req.Address != "" {
		companyDao.Address = req.Address
	}
	if req.LogoBase64 != "" {
		companyDao.LogoBase64 = req.LogoBase64
	}
	companyDao.UpdatedAt = time.Now()

	_, err = permission.Update(ctx, s.db.Conn, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("error updating company: %w", err)
	}

	return companyDaoToCompanyModel(*companyDao), nil
}

func (s *CompanyManagementServiceImpl) GetCompanies(ctx context.Context, id string) ([]*models.Company, error) {
	companiesDaos, err := dao.Companies(
		qm.InnerJoin("permissions p on p.company_id = companies.id"),
		qm.Where("p.user_id = ?", id),
	).All(ctx, s.db.Conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get Companies from database via join: %w", err)
	}

	var companies []*models.Company
	for _, companyDao := range companiesDaos {
		companies = append(companies, companyDaoToCompanyModel(*companyDao))
	}

	return companies, nil
}

func companyDaoToCompanyModel(companyDao dao.Company) *models.Company {
	return &models.Company{
		ID:         companyDao.ID,
		Name:       companyDao.Name,
		ContactID:  companyDao.ContactID,
		Address:    companyDao.Address,
		LogoBase64: companyDao.LogoBase64,
		CreatedAt:  companyDao.CreatedAt,
		UpdatedAt:  companyDao.UpdatedAt,
	}
}