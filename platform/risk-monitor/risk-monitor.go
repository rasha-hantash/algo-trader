package riskmonitor

import (
    "sync"
    "time"
    "context"
)

// RiskThreshold defines the risk parameters for a position
type RiskThreshold struct {
    Symbol          string
    OrderID         string
    MaxLoss         float64
    StopLoss        float64
    MaxPositionSize float64
    CurrentPosition float64
    EntryPrice      float64
    AccountRisk     float64
    LastUpdated     time.Time
}

// MarketData represents incoming market data
type MarketData struct {
    Symbol string
    Price  float64
    Volume float64
    Time   time.Time
}

// RiskMonitor manages risk monitoring for multiple positions
type RiskMonitor struct {
    // Concurrent map to store risk thresholds
    positions sync.Map
    // Channel for market data updates
    marketDataChan chan MarketData
    // Channel for risk alerts
    alertChan chan RiskAlert
    // Channel for position updates
    positionUpdateChan chan RiskThreshold
    // Context for cleanup
    ctx context.Context
    // Mutex for aggregate risk calculations
    mu sync.RWMutex
    // Aggregate account risk
    totalAccountRisk float64
    maxAccountRisk   float64
}

type RiskAlert struct {
    Symbol      string
    AlertType   string
    CurrentRisk float64
    Threshold   float64
    Message     string
    Timestamp   time.Time
}

func NewRiskMonitor(ctx context.Context, maxAccountRisk float64) *RiskMonitor {
    rm := &RiskMonitor{
        marketDataChan:     make(chan MarketData, 1000),
        alertChan:         make(chan RiskAlert, 100),
        positionUpdateChan: make(chan RiskThreshold, 100),
        ctx:               ctx,
        maxAccountRisk:    maxAccountRisk,
    }
    
    go rm.monitorRisk()
    return rm
}

// AddPosition adds a new position to monitor
func (rm *RiskMonitor) AddPosition(threshold RiskThreshold) {
    rm.positions.Store(threshold.Symbol, threshold)
    rm.positionUpdateChan <- threshold
}

// UpdatePosition updates an existing position
func (rm *RiskMonitor) UpdatePosition(symbol string, updateFn func(*RiskThreshold)) {
    if val, ok := rm.positions.Load(symbol); ok {
        threshold := val.(RiskThreshold)
        updateFn(&threshold)
        rm.positions.Store(symbol, threshold)
        rm.positionUpdateChan <- threshold
    }
}

// HandleMarketData processes incoming market data
func (rm *RiskMonitor) HandleMarketData(data MarketData) {
    rm.marketDataChan <- data
}

// monitorRisk is the main risk monitoring loop
func (rm *RiskMonitor) monitorRisk() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-rm.ctx.Done():
            return
            
        case data := <-rm.marketDataChan:
            rm.checkPositionRisk(data)
            
        case update := <-rm.positionUpdateChan:
            rm.updateAccountRisk(update)
            
        case <-ticker.C:
            rm.performPeriodicRiskCheck()
        }
    }
}

func (rm *RiskMonitor) checkPositionRisk(data MarketData) {
    if val, ok := rm.positions.Load(data.Symbol); ok {
        threshold := val.(RiskThreshold)
        
        // Calculate current P&L
        pnl := (data.Price - threshold.EntryPrice) * threshold.CurrentPosition
        riskPercentage := -pnl / threshold.MaxLoss * 100

        // Check stop loss
        if data.Price <= threshold.StopLoss {
            rm.alertChan <- RiskAlert{
                Symbol:    data.Symbol,
                AlertType: "STOP_LOSS_TRIGGERED",
                Message:   "Position hit stop loss",
                Timestamp: data.Time,
            }
        }

        // Check risk threshold
        if riskPercentage >= 80 { // Alert at 80% of max risk
            rm.alertChan <- RiskAlert{
                Symbol:      data.Symbol,
                AlertType:   "RISK_THRESHOLD_WARNING",
                CurrentRisk: riskPercentage,
                Threshold:   80,
                Message:     "Position approaching max risk threshold",
                Timestamp:   data.Time,
            }
        }
    }
}

func (rm *RiskMonitor) updateAccountRisk(update RiskThreshold) {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    // Recalculate total account risk
    var totalRisk float64
    rm.positions.Range(func(key, value interface{}) bool {
        pos := value.(RiskThreshold)
        totalRisk += pos.AccountRisk
        return true
    })

    rm.totalAccountRisk = totalRisk

    // Check if total risk exceeds maximum
    if rm.totalAccountRisk > rm.maxAccountRisk {
        rm.alertChan <- RiskAlert{
            AlertType:   "ACCOUNT_RISK_EXCEEDED",
            CurrentRisk: rm.totalAccountRisk,
            Threshold:   rm.maxAccountRisk,
            Message:     "Total account risk exceeds maximum",
            Timestamp:   time.Now(),
        }
    }
}

func (rm *RiskMonitor) performPeriodicRiskCheck() {
    now := time.Now()
    rm.positions.Range(func(key, value interface{}) bool {
        threshold := value.(RiskThreshold)
        
        // Check for stale positions
        if now.Sub(threshold.LastUpdated) > 5*time.Minute {
            rm.alertChan <- RiskAlert{
                Symbol:    threshold.Symbol,
                AlertType: "STALE_POSITION",
                Message:   "Position hasn't been updated in 5 minutes",
                Timestamp: now,
            }
        }
        return true
    })
}

// GetAlerts returns the channel for receiving risk alerts
func (rm *RiskMonitor) GetAlerts() <-chan RiskAlert {
    return rm.alertChan
}