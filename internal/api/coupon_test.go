package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListCoupons(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/coupon" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"couponId":"coupon-123","title":"Summer Sale","status":"RUNNING"},{"couponId":"coupon-456","title":"Winter Sale","status":"DRAFT"}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.ListCoupons(context.Background(), nil, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Coupons) != 2 {
		t.Errorf("expected 2 coupons, got %d", len(resp.Coupons))
	}
	if resp.Coupons[0].CouponID != "coupon-123" {
		t.Errorf("expected coupon-123, got %s", resp.Coupons[0].CouponID)
	}
	if resp.Coupons[0].Title != "Summer Sale" {
		t.Errorf("expected Summer Sale, got %s", resp.Coupons[0].Title)
	}
	if resp.Coupons[0].Status != "RUNNING" {
		t.Errorf("expected RUNNING, got %s", resp.Coupons[0].Status)
	}
}

func TestClient_ListCoupons_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		// Check query parameters
		if r.URL.Query().Get("status") != "RUNNING" {
			t.Errorf("expected status=RUNNING, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"couponId":"coupon-123","title":"Summer Sale","status":"RUNNING"}],"next":"cursor-abc"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.ListCoupons(context.Background(), []string{"RUNNING"}, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Coupons) != 1 {
		t.Errorf("expected 1 coupon, got %d", len(resp.Coupons))
	}
	if resp.Next != "cursor-abc" {
		t.Errorf("expected cursor-abc, got %s", resp.Next)
	}
}

func TestClient_CreateCoupon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/coupon" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req CreateCouponRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to parse request body: %v", err)
		}
		if req.Title != "Summer Sale" {
			t.Errorf("expected Summer Sale, got %s", req.Title)
		}
		if req.EndTimestamp != 1735689600000 {
			t.Errorf("expected 1735689600000, got %d", req.EndTimestamp)
		}
		if req.Reward == nil {
			t.Error("expected reward to be set")
		} else {
			if req.Reward.Type != "discount" {
				t.Errorf("expected discount, got %s", req.Reward.Type)
			}
			if req.Reward.PriceInfo == nil {
				t.Error("expected priceInfo to be set")
			} else if req.Reward.PriceInfo.FixedAmount != 500 {
				t.Errorf("expected 500, got %d", req.Reward.PriceInfo.FixedAmount)
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"couponId":"coupon-new-123"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	req := &CreateCouponRequest{
		Title:        "Summer Sale",
		EndTimestamp: 1735689600000,
		Reward: &CouponReward{
			Type: "discount",
			PriceInfo: &CouponPriceInfo{
				Type:        "fixed",
				FixedAmount: 500,
			},
		},
	}

	couponID, err := client.CreateCoupon(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if couponID != "coupon-new-123" {
		t.Errorf("expected coupon-new-123, got %s", couponID)
	}
}

func TestClient_GetCoupon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/coupon/coupon-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"couponId": "coupon-123",
			"title": "Summer Sale",
			"description": "Get 500 off",
			"status": "RUNNING",
			"startTimestamp": 1704067200000,
			"endTimestamp": 1735689600000,
			"timezone": "Asia/Tokyo",
			"reward": {
				"type": "discount",
				"priceInfo": {
					"type": "fixed",
					"fixedAmount": 500
				}
			}
		}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	coupon, err := client.GetCoupon(context.Background(), "coupon-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if coupon.CouponID != "coupon-123" {
		t.Errorf("expected coupon-123, got %s", coupon.CouponID)
	}
	if coupon.Title != "Summer Sale" {
		t.Errorf("expected Summer Sale, got %s", coupon.Title)
	}
	if coupon.Description != "Get 500 off" {
		t.Errorf("expected Get 500 off, got %s", coupon.Description)
	}
	if coupon.Status != "RUNNING" {
		t.Errorf("expected RUNNING, got %s", coupon.Status)
	}
	if coupon.Timezone != "Asia/Tokyo" {
		t.Errorf("expected Asia/Tokyo, got %s", coupon.Timezone)
	}
	if coupon.Reward == nil {
		t.Error("expected reward to be set")
	} else {
		if coupon.Reward.Type != "discount" {
			t.Errorf("expected discount, got %s", coupon.Reward.Type)
		}
		if coupon.Reward.PriceInfo == nil {
			t.Error("expected priceInfo to be set")
		} else if coupon.Reward.PriceInfo.FixedAmount != 500 {
			t.Errorf("expected 500, got %d", coupon.Reward.PriceInfo.FixedAmount)
		}
	}
}

func TestClient_CloseCoupon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/v2/bot/coupon/coupon-123/close" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	err := client.CloseCoupon(context.Background(), "coupon-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ListCoupons_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.ListCoupons(context.Background(), nil, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Coupons) != 0 {
		t.Errorf("expected 0 coupons, got %d", len(resp.Coupons))
	}
}

func TestClient_ListCoupons_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("start") != "cursor-xyz" {
			t.Errorf("expected start=cursor-xyz, got %s", r.URL.Query().Get("start"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"couponId":"coupon-789","title":"Page 2"}]}`))
	}))
	defer server.Close()

	client := NewClient("test-token", false, false)
	client.baseURL = server.URL

	resp, err := client.ListCoupons(context.Background(), nil, 0, "cursor-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Coupons) != 1 {
		t.Errorf("expected 1 coupon, got %d", len(resp.Coupons))
	}
	if resp.Coupons[0].CouponID != "coupon-789" {
		t.Errorf("expected coupon-789, got %s", resp.Coupons[0].CouponID)
	}
}
