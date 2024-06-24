package controllers

import (
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type VesselController struct {
	// Dependent services
}

func NewVesselController() *VesselController {
	return &VesselController{
		// Inject services
	}
}

func (r *VesselController) GetVessel(c *gin.Context) {
	// TODO: Change if want to use another model here
	var data []models.Kapal

	// Fetch users data from the database
	database.DB.Find(&data)

	draw, _ := strconv.Atoi(c.Query("draw"))
	start, _ := strconv.Atoi(c.Query("start"))
	length, _ := strconv.Atoi(c.Query("length"))
	search := c.Query("search[value]") // Get search value from DataTables request

	// Filter data based on search query
	// TODO: Change if want to use another model here
	filteredData := make([]models.Kapal, 0)
	for _, vessel := range data {
		// Iterate over each field of the user struct
		v := reflect.ValueOf(vessel)
		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i).Interface()
			// Check if the field value contains the search query
			if strings.Contains(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(search)) {
				// If any field contains the search query, add the user to filtered data and break the loop
				filteredData = append(filteredData, vessel)
				break
			}
		}
	}

	totalRecords := len(filteredData)

	// Sort data based on order specified by DataTables
	orderColumnIndex, _ := strconv.Atoi(c.Query("order[0][column]"))
	orderDirection := c.Query("order[0][dir]")

	switch orderColumnIndex {
	case 0: // Assuming sorting is based on ID column
		if orderDirection == "asc" {
			sort.Slice(filteredData, func(i, j int) bool {
				return filteredData[i].CallSign < filteredData[j].CallSign
			})
		} else {
			sort.Slice(filteredData, func(i, j int) bool {
				return filteredData[i].CallSign > filteredData[j].CallSign
			})
		}
		// Add cases for other columns if needed
	}

	// Slice the data to return only the portion needed for the current page
	end := start + length
	if end > totalRecords {
		end = totalRecords
	}
	pagedData := filteredData[start:end]

	// Send JSON response to DataTables
	c.JSON(http.StatusOK, gin.H{
		"draw":            draw,
		"recordsTotal":    len(data),
		"recordsFiltered": totalRecords,
		"data":            pagedData,
	})

}
