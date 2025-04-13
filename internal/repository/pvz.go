package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kosttiik/pvz-service/internal/models"
)

type PVZRepository struct {
	db *pgxpool.Pool
}

type GetPVZFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	Limit     int
}

type PVZandReceptions struct {
	PVZ        models.PVZ
	Receptions []ReceptionAndProducts
}

type ReceptionAndProducts struct {
	Reception models.Reception
	Products  []models.Product
}

func NewPVZRepository(db *pgxpool.Pool) *PVZRepository {
	return &PVZRepository{db: db}
}

func (r *PVZRepository) GetPVZ(ctx context.Context, filter GetPVZFilter) ([]PVZandReceptions, error) {
	var conditions []string
	var args []any
	argPos := 1

	// Запрос с With для фильтрации по датам
	query := `
        WITH filtered_pvz AS (
            SELECT DISTINCT p.* 
            FROM pvz p
            LEFT JOIN reception r ON p.id = r.pvz_id
            WHERE 1=1
    `

	// Добавляем условия фильтрации по датам
	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("r.date_time >= $%d", argPos))
		args = append(args, filter.StartDate)
		argPos++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("r.date_time <= $%d", argPos))
		args = append(args, filter.EndDate)
		argPos++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += fmt.Sprintf(`
        )
        SELECT p.id, p.registration_date, p.city,
               r.id, r.date_time, r.status,
               pr.id, pr.date_time, pr.type
        FROM filtered_pvz p
        LEFT JOIN reception r ON p.id = r.pvz_id
        LEFT JOIN product pr ON r.id = pr.reception_id
        ORDER BY p.registration_date DESC
        LIMIT $%d OFFSET $%d
    `, argPos, argPos+1)

	// пагинация
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query PVZs: %w", err)
	}
	defer rows.Close()

	pvzMap := make(map[string]*PVZandReceptions)
	for rows.Next() {
		var pvz models.PVZ
		var receptionID, receptionDateTime, receptionStatus sql.NullString
		var productID, productDateTime, productType sql.NullString

		err := rows.Scan(
			&pvz.ID, &pvz.RegistrationDate, &pvz.City,
			&receptionID, &receptionDateTime, &receptionStatus,
			&productID, &productDateTime, &productType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Добавляем ПВЗ в map если его там еще нет
		if _, exists := pvzMap[pvz.ID.String()]; !exists {
			pvzMap[pvz.ID.String()] = &PVZandReceptions{
				PVZ: pvz,
			}
		}

		// Если есть приемка, добавляем ее
		if receptionID.Valid {
			reception := models.Reception{
				ID:       uuid.MustParse(receptionID.String),
				DateTime: parseTime(receptionDateTime.String),
				Status:   models.ReceptionStatus(receptionStatus.String),
				PvzID:    pvz.ID.String(),
			}

			// Ищем существующую приемку или создаем новую
			var found bool
			for i := range pvzMap[pvz.ID.String()].Receptions {
				if pvzMap[pvz.ID.String()].Receptions[i].Reception.ID == reception.ID {
					found = true
					// Если есть товар, добавляем его
					if productID.Valid {
						product := models.Product{
							ID:          uuid.MustParse(productID.String),
							DateTime:    parseTime(productDateTime.String),
							Type:        productType.String,
							ReceptionID: reception.ID.String(),
						}
						pvzMap[pvz.ID.String()].Receptions[i].Products = append(
							pvzMap[pvz.ID.String()].Receptions[i].Products,
							product,
						)
					}
					break
				}
			}

			if !found {
				receptionWithProducts := ReceptionAndProducts{Reception: reception}
				if productID.Valid {
					product := models.Product{
						ID:          uuid.MustParse(productID.String),
						DateTime:    parseTime(productDateTime.String),
						Type:        productType.String,
						ReceptionID: reception.ID.String(),
					}
					receptionWithProducts.Products = []models.Product{product}
				}
				pvzMap[pvz.ID.String()].Receptions = append(
					pvzMap[pvz.ID.String()].Receptions,
					receptionWithProducts,
				)
			}
		}
	}

	// Конвертируем map в слайс для ответа
	result := make([]PVZandReceptions, 0, len(pvzMap))
	for _, pvz := range pvzMap {
		result = append(result, *pvz)
	}

	return result, nil
}

func parseTime(t string) time.Time {
	parsed, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
