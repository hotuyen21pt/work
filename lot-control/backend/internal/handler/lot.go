package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lot-control/internal/middleware"
	"lot-control/internal/models"
)

type LotHandler struct {
	db *sql.DB
}

func NewLotHandler(db *sql.DB) *LotHandler {
	return &LotHandler{db: db}
}

func (h *LotHandler) List(c *gin.Context) {
	skuID := c.Query("sku_id")
	if skuID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku_id là bắt buộc"})
		return
	}

	rows, err := h.db.Query(`
		SELECT l.id, l.sku_id, l.lot_number, l.manufacture_date, l.expiry_date, l.qty,
		       l.branch, l.counted_by, COALESCE(u.full_name, ''), l.counted_at, l.notes
		FROM lots l
		LEFT JOIN users u ON u.id = l.counted_by
		WHERE l.sku_id = ? ORDER BY l.lot_number ASC`, skuID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	lots := []models.Lot{}
	for rows.Next() {
		var lot models.Lot
		if err := rows.Scan(
			&lot.ID, &lot.SKUID, &lot.LotNumber, &lot.ManufactureDate, &lot.ExpiryDate,
			&lot.Qty, &lot.Branch, &lot.CountedByID, &lot.CountedByName,
			&lot.CountedAt, &lot.Notes,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		lots = append(lots, lot)
	}

	c.JSON(http.StatusOK, lots)
}

func (h *LotHandler) Upsert(c *gin.Context) {
	claims := middleware.GetClaims(c)

	var req models.UpsertLotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branch := req.Branch
	if branch == "" && claims != nil {
		branch = claims.Branch
	}

	var skuExists int
	h.db.QueryRow(`SELECT COUNT(*) FROM skus WHERE id=?`, req.SKUID).Scan(&skuExists)
	if skuExists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy SKU"})
		return
	}

	var userID *int64
	if claims != nil {
		uid := claims.UserID
		userID = &uid
	}

	_, err := h.db.Exec(`
		INSERT INTO lots (sku_id, lot_number, manufacture_date, expiry_date, qty, branch, counted_by, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(sku_id, lot_number) DO UPDATE SET
			manufacture_date = excluded.manufacture_date,
			expiry_date      = excluded.expiry_date,
			qty              = excluded.qty,
			branch           = excluded.branch,
			counted_by       = excluded.counted_by,
			counted_at       = CURRENT_TIMESTAMP,
			notes            = excluded.notes`,
		req.SKUID, req.LotNumber, req.ManufactureDate, req.ExpiryDate,
		req.Qty, branch, userID, req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.db.Exec(`UPDATE skus SET updated_at=CURRENT_TIMESTAMP WHERE id=?`, req.SKUID)

	var lot models.Lot
	h.db.QueryRow(`
		SELECT l.id, l.sku_id, l.lot_number, l.manufacture_date, l.expiry_date, l.qty,
		       l.branch, l.counted_by, COALESCE(u.full_name, ''), l.counted_at, l.notes
		FROM lots l LEFT JOIN users u ON u.id = l.counted_by
		WHERE l.sku_id = ? AND l.lot_number = ?`, req.SKUID, req.LotNumber,
	).Scan(
		&lot.ID, &lot.SKUID, &lot.LotNumber, &lot.ManufactureDate, &lot.ExpiryDate,
		&lot.Qty, &lot.Branch, &lot.CountedByID, &lot.CountedByName, &lot.CountedAt, &lot.Notes,
	)

	c.JSON(http.StatusOK, lot)
}

func (h *LotHandler) Update(c *gin.Context) {
	claims := middleware.GetClaims(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id không hợp lệ"})
		return
	}

	var req models.UpdateLotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *int64
	if claims != nil {
		uid := claims.UserID
		userID = &uid
	}

	res, err := h.db.Exec(`
		UPDATE lots SET manufacture_date=?, expiry_date=?, qty=?, notes=?,
		               counted_by=?, counted_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		req.ManufactureDate, req.ExpiryDate, req.Qty, req.Notes, userID, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy lô"})
		return
	}

	var lot models.Lot
	h.db.QueryRow(`
		SELECT l.id, l.sku_id, l.lot_number, l.manufacture_date, l.expiry_date, l.qty,
		       l.branch, l.counted_by, COALESCE(u.full_name, ''), l.counted_at, l.notes
		FROM lots l LEFT JOIN users u ON u.id = l.counted_by WHERE l.id=?`, id,
	).Scan(
		&lot.ID, &lot.SKUID, &lot.LotNumber, &lot.ManufactureDate, &lot.ExpiryDate,
		&lot.Qty, &lot.Branch, &lot.CountedByID, &lot.CountedByName, &lot.CountedAt, &lot.Notes,
	)

	c.JSON(http.StatusOK, lot)
}

func (h *LotHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id không hợp lệ"})
		return
	}

	res, err := h.db.Exec(`DELETE FROM lots WHERE id=?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy lô"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã xóa"})
}
