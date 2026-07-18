package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopcaovanson/auth-service/internal/domain"
	"github.com/shopcaovanson/auth-service/internal/repository"
	"github.com/shopcaovanson/auth-service/internal/storage"
	"github.com/xuri/excelize/v2"
)

var vietnamTZ = mustLoadLocation("Asia/Ho_Chi_Minh")

func VietnamLocation() *time.Location {
	return vietnamTZ
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.FixedZone("ICT", 7*3600)
	}
	return loc
}

type LoginHistoryService struct {
	history repository.LoginHistoryRepository
	reports repository.LoginReportRepository
	store   *storage.AvatarStorage
}

func NewLoginHistoryService(
	history repository.LoginHistoryRepository,
	reports repository.LoginReportRepository,
	store *storage.AvatarStorage,
) *LoginHistoryService {
	return &LoginHistoryService{history: history, reports: reports, store: store}
}

func (s *LoginHistoryService) List(ctx context.Context, filter domain.LoginHistoryFilter) (*domain.LoginHistoryResult, error) {
	return s.history.List(ctx, filter)
}

func (s *LoginHistoryService) ListReports(ctx context.Context, page, limit int, from, to *time.Time) (*domain.LoginReportListResult, error) {
	return s.reports.List(ctx, page, limit, from, to)
}

func (s *LoginHistoryService) GetReport(ctx context.Context, id uuid.UUID) (*domain.LoginReport, error) {
	return s.reports.GetByID(ctx, id)
}

func (s *LoginHistoryService) OpenReportFile(ctx context.Context, id uuid.UUID) (*storage.Object, *domain.LoginReport, error) {
	report, err := s.reports.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if report == nil {
		return nil, nil, fmt.Errorf("report not found")
	}
	if report.Status != "ready" {
		return nil, nil, fmt.Errorf("report is not ready")
	}
	obj, err := s.store.OpenObject(ctx, report.ObjectKey)
	if err != nil {
		return nil, nil, err
	}
	return obj, report, nil
}

// DayBoundsVN returns [start, end) for a calendar day in Vietnam timezone, as UTC instants.
func DayBoundsVN(day time.Time) (from, to time.Time) {
	local := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, vietnamTZ)
	return local.UTC(), local.Add(24 * time.Hour).UTC()
}

func YesterdayVN(now time.Time) time.Time {
	local := now.In(vietnamTZ).AddDate(0, 0, -1)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, vietnamTZ)
}

// GenerateDailyReport builds Excel for the given VN calendar day, uploads to MinIO, upserts login_reports.
func (s *LoginHistoryService) GenerateDailyReport(ctx context.Context, reportDay time.Time) (*domain.LoginReport, error) {
	if !s.store.Enabled() {
		return nil, storage.ErrStorageDisabled
	}

	day := time.Date(reportDay.Year(), reportDay.Month(), reportDay.Day(), 0, 0, 0, 0, vietnamTZ)
	from, to := DayBoundsVN(day)
	dateStr := day.Format("2006-01-02")

	rows, err := s.history.ListByDateRange(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("list login history: %w", err)
	}
	total, successCount, failureCount, uniqueUsers, err := s.history.StatsByDateRange(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("stats: %w", err)
	}

	xlsx, err := buildLoginExcel(rows, dateStr, total, successCount, failureCount, uniqueUsers)
	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("login-history-%s.xlsx", dateStr)
	objectKey := storage.LoginReportObjectKey(dateStr, fileName)
	if err := s.store.UploadBytes(ctx, objectKey, xlsx, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	report := &domain.LoginReport{
		ID:           uuid.New(),
		ReportDate:   day,
		ObjectKey:    objectKey,
		FileName:     fileName,
		FileSize:     int64(len(xlsx)),
		TotalLogins:  total,
		SuccessCount: successCount,
		FailureCount: failureCount,
		UniqueUsers:  uniqueUsers,
		Status:       "ready",
		CreatedAt:    now,
	}
	if err := s.reports.Upsert(ctx, report); err != nil {
		return nil, fmt.Errorf("save report row: %w", err)
	}
	return report, nil
}

func buildLoginExcel(rows []domain.LoginHistory, dateStr string, total, success, failure, uniqueUsers int) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	summary := "Tong_hop"
	detail := "Chi_tiet"
	_ = f.SetSheetName("Sheet1", summary)
	_, _ = f.NewSheet(detail)

	_ = f.SetCellValue(summary, "A1", "Bao cao lich su dang nhap")
	_ = f.SetCellValue(summary, "A2", "Ngay")
	_ = f.SetCellValue(summary, "B2", dateStr)
	_ = f.SetCellValue(summary, "A3", "Tong so lan dang nhap")
	_ = f.SetCellValue(summary, "B3", total)
	_ = f.SetCellValue(summary, "A4", "Thanh cong")
	_ = f.SetCellValue(summary, "B4", success)
	_ = f.SetCellValue(summary, "A5", "That bai")
	_ = f.SetCellValue(summary, "B5", failure)
	_ = f.SetCellValue(summary, "A6", "So user thanh cong (unique)")
	_ = f.SetCellValue(summary, "B6", uniqueUsers)

	headers := []string{
		"Thoi gian (VN)", "Email", "Ho ten", "Ket qua", "Ly do that bai",
		"IP", "Thiet bi", "Trinh duyet", "He dieu hanh", "User-Agent", "User ID",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(detail, cell, h)
	}

	for i, row := range rows {
		r := i + 2
		status := "That bai"
		if row.Success {
			status = "Thanh cong"
		}
		fullName := ""
		if row.FullName != nil {
			fullName = *row.FullName
		}
		reason := ""
		if row.FailureReason != nil {
			reason = *row.FailureReason
		}
		ip := strPtr(row.IPAddress)
		device := strPtr(row.Device)
		browser := strPtr(row.Browser)
		osName := strPtr(row.OS)
		ua := strPtr(row.UserAgent)
		userID := ""
		if row.UserID != nil {
			userID = row.UserID.String()
		}
		values := []interface{}{
			row.CreatedAt.In(vietnamTZ).Format("2006-01-02 15:04:05"),
			row.Email,
			fullName,
			status,
			reason,
			ip,
			device,
			browser,
			osName,
			ua,
			userID,
		}
		for c, v := range values {
			cell, _ := excelize.CoordinatesToCellName(c+1, r)
			_ = f.SetCellValue(detail, cell, v)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func strPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
