// Copyright 2020
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Code generated by gen-accessors; DO NOT EDIT.
package api

// HasNotificationChannels checks if AutoscalingRequest has any NotificationChannels.
func (a *AutoscalingRequest) HasNotificationChannels() bool {
	if a == nil || a.NotificationChannels == nil {
		return false
	}
	if len(a.NotificationChannels) == 0 {
		return false
	}
	return true
}

// GetActionType returns the ActionType field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetActionType() string {
	if f == nil || f.ActionType == nil {
		return ""
	}
	return *f.ActionType
}

// GetAppID returns the AppID field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetAppID() string {
	if f == nil || f.AppID == nil {
		return ""
	}
	return *f.AppID
}

// GetForecastPeriod returns the ForecastPeriod field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetForecastPeriod() int {
	if f == nil || f.ForecastPeriod == nil {
		return 0
	}
	return *f.ForecastPeriod
}

// GetID returns the ID field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetID() string {
	if f == nil || f.ID == nil {
		return ""
	}
	return *f.ID
}

// GetIsActive returns the IsActive field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetIsActive() bool {
	if f == nil || f.IsActive == nil {
		return false
	}
	return *f.IsActive
}

// GetMaxQuantity returns the MaxQuantity field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetMaxQuantity() int {
	if f == nil || f.MaxQuantity == nil {
		return 0
	}
	return *f.MaxQuantity
}

// GetMetricUUID returns the MetricUUID field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetMetricUUID() string {
	if f == nil || f.MetricUUID == nil {
		return ""
	}
	return *f.MetricUUID
}

// GetMinQuantity returns the MinQuantity field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetMinQuantity() int {
	if f == nil || f.MinQuantity == nil {
		return 0
	}
	return *f.MinQuantity
}

// GetName returns the Name field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetName() string {
	if f == nil || f.Name == nil {
		return ""
	}
	return *f.Name
}

// HasNotificationChannels checks if FormationMonitor has any NotificationChannels.
func (f *FormationMonitor) HasNotificationChannels() bool {
	if f == nil || f.NotificationChannels == nil {
		return false
	}
	if len(f.NotificationChannels) == 0 {
		return false
	}
	return true
}

// GetNotificationPeriod returns the NotificationPeriod field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetNotificationPeriod() int {
	if f == nil || f.NotificationPeriod == nil {
		return 0
	}
	return *f.NotificationPeriod
}

// GetOperation returns the Operation field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetOperation() string {
	if f == nil || f.Operation == nil {
		return ""
	}
	return *f.Operation
}

// GetPeriod returns the Period field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetPeriod() int {
	if f == nil || f.Period == nil {
		return 0
	}
	return *f.Period
}

// GetProcessType returns the ProcessType field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetProcessType() string {
	if f == nil || f.ProcessType == nil {
		return ""
	}
	return *f.ProcessType
}

// GetState returns the State field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetState() string {
	if f == nil || f.State == nil {
		return ""
	}
	return *f.State
}

// GetValue returns the Value field if it's non-nil, zero value otherwise.
func (f *FormationMonitor) GetValue() int {
	if f == nil || f.Value == nil {
		return 0
	}
	return *f.Value
}
