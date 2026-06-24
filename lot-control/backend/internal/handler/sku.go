package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lot-control/internal/models"
)

type SKUHandler struct {
	db *sql.DB
}

func NewSKUHandler(db *sql.DB) *SKUHandler {
	return &SKUHandler{db: db}
}

func (h *SKUHandler) List(c *gin.Context) {
	q := c.Query("q")

	query := `
		SELECT s.id, s.sku_code, s.name, s.unit, s.created_at, s.updated_at,
		       COALESCE(SUM(l.qty), 0), COUNT(l.id)
		FROM skus s
		LEFT JOIN lots l ON l.sku_id = s.id`

	args := []interface{}{}
	if q != "" {
		query += ` WHERE s.sku_code LIKE ? OR s.name LIKE ?`
		like := "%" + q + "%"
		args = append(args, like, like)
	}
	query += ` GROUP BY s.id ORDER BY s.updated_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	skus := []models.SKU{}
	for rows.Next() {
		var s models.SKU
		if err := rows.Scan(&s.ID, &s.SKUCode, &s.Name, &s.Unit, &s.CreatedAt, &s.UpdatedAt, &s.TotalQty, &s.LotCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		skus = append(skus, s)
	}

	c.JSON(http.StatusOK, skus)
}

func (h *SKUHandler) Create(c *gin.Context) {
	var req models.CreateSKURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Unit == "" {
		req.Unit = "cái"
	}

	res, err := h.db.Exec(
		`INSERT INTO skus (sku_code, name, unit) VALUES (?, ?, ?)`,
		req.SKUCode, req.Name, req.Unit,
	)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "mã SKU đã tồn tại"})
		return
	}

	id, _ := res.LastInsertId()
	var sku models.SKU
	h.db.QueryRow(
		`SELECT id, sku_code, name, unit, created_at, updated_at FROM skus WHERE id = ?`, id,
	).Scan(&sku.ID, &sku.SKUCode, &sku.Name, &sku.Unit, &sku.CreatedAt, &sku.UpdatedAt)

	c.JSON(http.StatusCreated, sku)
}

func (h *SKUHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id không hợp lệ"})
		return
	}

	var sku models.SKU
	err = h.db.QueryRow(`
		SELECT s.id, s.sku_code, s.name, s.unit, s.created_at, s.updated_at,
		       COALESCE(SUM(l.qty), 0), COUNT(l.id)
		FROM skus s LEFT JOIN lots l ON l.sku_id = s.id
		WHERE s.id = ? GROUP BY s.id`, id,
	).Scan(&sku.ID, &sku.SKUCode, &sku.Name, &sku.Unit, &sku.CreatedAt, &sku.UpdatedAt, &sku.TotalQty, &sku.LotCount)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy SKU"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := h.db.Query(`
		SELECT l.id, l.lot_number, l.manufacture_date, l.expiry_date, l.qty,
		       l.branch, l.counted_by, COALESCE(u.full_name, ''), l.counted_at, l.notes
		FROM lots l
		LEFT JOIN users u ON u.id = l.counted_by
		WHERE l.sku_id = ? ORDER BY l.lot_number ASC`, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	sku.Lots = []models.Lot{}
	for rows.Next() {
		var lot models.Lot
		lot.SKUID = id
		if err := rows.Scan(
			&lot.ID, &lot.LotNumber, &lot.ManufactureDate, &lot.ExpiryDate,
			&lot.Qty, &lot.Branch, &lot.CountedByID, &lot.CountedByName,
			&lot.CountedAt, &lot.Notes,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		sku.Lots = append(sku.Lots, lot)
	}

	c.JSON(http.StatusOK, sku)
}

func (h *SKUHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id không hợp lệ"})
		return
	}

	var req models.UpdateSKURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Unit == "" {
		req.Unit = "cái"
	}

	_, err = h.db.Exec(
		`UPDATE skus SET name=?, unit=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		req.Name, req.Unit, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã cập nhật"})
}

func (h *SKUHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id không hợp lệ"})
		return
	}

	res, err := h.db.Exec(`DELETE FROM skus WHERE id=?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy SKU"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã xóa"})
}
