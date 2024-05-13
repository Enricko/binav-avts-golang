package controllers

import (
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

)

type PendudukController struct {
	// Dependent services
}

func NewPendudukController() *PendudukController {
	return &PendudukController{
		// Inject services
	}
}
func (r *PendudukController) Index(c *gin.Context) {
	// Data to pass to the index.html template

	var users []models.User

	// Retrieve all users from the database
	err := database.DB.Find(&users).Error

	if err != nil {
		fmt.Println(err)
	}

	app := gin.H{
		"title": "Dashboard Panel",
		"users": users,
	}

	c.HTML(http.StatusOK, "penduduk.html", app)
}

func (r *PendudukController) List(start, length int) ([]map[string]interface{}, int) {
    data := [][]string{
        {"John",
         "25", 
        },
    }
    var formattedData []map[string]interface{}
    for _, row := range data {
        rowData := make(map[string]interface{})
        rowData["nama"] = row[0]
        rowData["usia"], _ = strconv.Atoi(row[1])
        formattedData = append(formattedData, rowData)
    }

    // Menghitung jumlah total record
    totalRecords := len(formattedData)

    // Memastikan start tidak melebihi kapasitas formattedData
    if start >= totalRecords {
        return nil, totalRecords
    }

    // Memastikan length tidak melebihi kapasitas formattedData dari posisi start
    end := start + length
    if end > totalRecords {
        end = totalRecords
    }

    // Mengembalikan data yang diminta dan jumlah total record
    return formattedData[start:end], totalRecords
}


func (r *PendudukController) DataPenduduk(c *gin.Context) {
	draw, _ := strconv.Atoi(c.Query("draw"))
	start, _ := strconv.Atoi(c.Query("start"))
	length, _ := strconv.Atoi(c.Query("length"))

	// Mendapatkan data sesuai dengan start dan length yang diterima dari DataTables
	data, totalRecords := r.List(start, length)

	// Kirimkan respons JSON ke DataTables.
	c.JSON(http.StatusOK, gin.H{
		"draw":            draw,
		"recordsTotal":    totalRecords, // Jumlah total record
		"recordsFiltered": totalRecords, // Jumlah record yang difilter (dalam kasus ini, sama dengan jumlah total record)
		"data":            data,         // Data yang akan ditampilkan pada halaman yang diminta
	})
}
