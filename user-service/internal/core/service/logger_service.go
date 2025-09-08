package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"user-service/internal/core/domain/entity"

	"github.com/labstack/gommon/log"
)

// OAuthFileLogger handles OAuth activity logging to files
type OAuthFileLogger struct {
	logDir string
}

// NewOAuthFileLogger creates a new file logger
func NewOAuthFileLogger(logDir string) *OAuthFileLogger {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Errorf("[OAuthFileLogger] Failed to create log directory: %v", err)
	}

	return &OAuthFileLogger{
		logDir: logDir,
	}
}

// LogOAuthActivity logs OAuth activity to file
func (l *OAuthFileLogger) LogOAuthActivity(ctx context.Context, userID int64, provider, action, status, errorMsg, userAgent, ipAddress string) {
	activity := &entity.OAuthActivityLog{
		UserID:    userID,
		Provider:  provider,
		Action:    action,
		Status:    status,
		ErrorMsg:  errorMsg,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
	}

	// Log to file asynchronously to avoid blocking
	go l.writeToFile(activity)
}

// writeToFile writes the activity log to file
func (l *OAuthFileLogger) writeToFile(activity *entity.OAuthActivityLog) {
	// Generate filename based on date (daily rotation)
	date := activity.CreatedAt.Format("2006-01-02")
	filename := fmt.Sprintf("oauth_%s.log", date)
	filepath := filepath.Join(l.logDir, filename)

	// Prepare log entry
	logEntry := map[string]interface{}{
		"timestamp":  activity.CreatedAt.Format(time.RFC3339),
		"user_id":    activity.UserID,
		"provider":   activity.Provider,
		"action":     activity.Action,
		"status":     activity.Status,
		"ip_address": activity.IPAddress,
		"user_agent": activity.UserAgent,
		"error_msg":  activity.ErrorMsg,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Errorf("[OAuthFileLogger] Failed to marshal log entry: %v", err)
		return
	}

	// Open file for appending (create if not exists)
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Errorf("[OAuthFileLogger] Failed to open log file: %v", err)
		return
	}
	defer file.Close()

	// Write JSON line to file
	if _, err := file.WriteString(string(jsonData) + "\n"); err != nil {
		log.Errorf("[OAuthFileLogger] Failed to write to log file: %v", err)
		return
	}
}

// GetOAuthActivitiesFromFile reads OAuth activities from log files
func (l *OAuthFileLogger) GetOAuthActivitiesFromFile(userID int64, days int, limit int) ([]*entity.OAuthActivityLog, error) {
	var activities []*entity.OAuthActivityLog

	// Read files for the last N days
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		filename := fmt.Sprintf("oauth_%s.log", date)
		filepath := filepath.Join(l.logDir, filename)

		fileActivities, err := l.readActivitiesFromFile(filepath, userID)
		if err != nil {
			// Skip if file doesn't exist or can't be read
			continue
		}

		activities = append(activities, fileActivities...)

		// Stop if we have enough activities
		if len(activities) >= limit {
			break
		}
	}

	// Sort by timestamp descending and limit results
	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}

// readActivitiesFromFile reads activities from a specific log file
func (l *OAuthFileLogger) readActivitiesFromFile(filepath string, userID int64) ([]*entity.OAuthActivityLog, error) {
	var activities []*entity.OAuthActivityLog

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return activities, nil
	}

	// Read file content
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Parse each line as JSON
	lines := string(content)
	for _, line := range splitLines(lines) {
		if line == "" {
			continue
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			continue // Skip malformed lines
		}

		// Filter by user ID if specified
		if userID > 0 {
			if logUserID, ok := logEntry["user_id"].(float64); !ok || int64(logUserID) != userID {
				continue
			}
		}

		// Convert to OAuthActivityLog
		activity := &entity.OAuthActivityLog{}

		if userIDVal, ok := logEntry["user_id"].(float64); ok {
			activity.UserID = int64(userIDVal)
		}

		if provider, ok := logEntry["provider"].(string); ok {
			activity.Provider = provider
		}

		if action, ok := logEntry["action"].(string); ok {
			activity.Action = action
		}

		if status, ok := logEntry["status"].(string); ok {
			activity.Status = status
		}

		if ipAddress, ok := logEntry["ip_address"].(string); ok {
			activity.IPAddress = ipAddress
		}

		if userAgent, ok := logEntry["user_agent"].(string); ok {
			activity.UserAgent = userAgent
		}

		if errorMsg, ok := logEntry["error_msg"].(string); ok {
			activity.ErrorMsg = errorMsg
		}

		if timestamp, ok := logEntry["timestamp"].(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, timestamp); err == nil {
				activity.CreatedAt = parsedTime
			}
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// CleanupOldLogs removes log files older than specified days
func (l *OAuthFileLogger) CleanupOldLogs(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Read directory
	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return err
	}

	// Check each file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Check if it's an OAuth log file
		if len(filename) < 16 || filename[:6] != "oauth_" || filename[len(filename)-4:] != ".log" {
			continue
		}

		// Extract date from filename (oauth_2024-01-15.log)
		dateStr := filename[6 : len(filename)-4]
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		// Delete if older than cutoff
		if fileDate.Before(cutoffDate) {
			filepath := filepath.Join(l.logDir, filename)
			if err := os.Remove(filepath); err != nil {
				log.Errorf("[OAuthFileLogger] Failed to delete old log file %s: %v", filename, err)
			} else {
				log.Infof("[OAuthFileLogger] Deleted old log file: %s", filename)
			}
		}
	}

	return nil
}

// splitLines splits content into lines
func splitLines(content string) []string {
	var lines []string
	var currentLine string

	for _, char := range content {
		if char == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(char)
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// logOAuthActivity - standalone function that uses file logger
func logOAuthActivity(ctx context.Context, userID int64, provider, action, status, errorMsg, userAgent, ipAddress string) {
	// Initialize file logger (you can make this global or inject as dependency)
	logger := NewOAuthFileLogger("./logs")
	logger.LogOAuthActivity(ctx, userID, provider, action, status, errorMsg, userAgent, ipAddress)
}
