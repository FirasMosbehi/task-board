// Package main provides HTTP route handlers for the TaskBoard API,
// including task creation, updates, retrieval, and deletion.
package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// getTasks returns all tasks ordered by creation date (newest first).
func getTasks(c *gin.Context) {
	var tasks []Task

	err := TrackDBOperation(c.Request.Context(), "query_all_tasks", func() error {
		return DB.Order("created_at desc").Find(&tasks).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tasks"})
		return
	}

	// Update metrics after successful retrieval
	go UpdateTaskMetrics(c.Request.Context())

	c.JSON(http.StatusOK, tasks)
}

// CreateTaskInput represents the expected payload for creating a new task.
type CreateTaskInput struct {
	Title string `json:"title" binding:"required,min=1,max=200"`
}

// createTask handles the creation of a new task.
func createTask(c *gin.Context) {
	var input CreateTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	task := Task{
		Title:     input.Title,
		Completed: false,
	}

	err := TrackDBOperation(c.Request.Context(), "create_task", func() error {
		return DB.Create(&task).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	// Update metrics after successful creation
	go UpdateTaskMetrics(c.Request.Context())

	c.JSON(http.StatusCreated, task)
}

// UpdateTaskInput represents the fields that can be updated in a task.
type UpdateTaskInput struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

// updateTask handles updates to an existing task.
func updateTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var task Task
	err = TrackDBOperation(c.Request.Context(), "find_task", func() error {
		return DB.First(&task, id).Error
	})
	
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	var input UpdateTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	
	if input.Title != nil {
		task.Title = *input.Title
	}

	if input.Completed != nil {
		task.Completed = *input.Completed
	}

	err = TrackDBOperation(c.Request.Context(), "update_task", func() error {
		return DB.Save(&task).Error
	})
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update task"})
		return
	}
	
	// Update metrics after successful update
	go UpdateTaskMetrics(c.Request.Context())
	
	c.JSON(http.StatusOK, task)
}

// deleteTask deletes a task by ID.
func deleteTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = TrackDBOperation(c.Request.Context(), "delete_task", func() error {
		return DB.Delete(&Task{}, id).Error
	})
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete task"})
		return
	}
	
	// Update metrics after successful deletion
	go UpdateTaskMetrics(c.Request.Context())

	c.Status(http.StatusNoContent)
}