package jsonkit

import (
	"testing"
)

// 测试用json数据
var demoJson = `{
    "id": "user12345",
    "active": true,
    "age": 32,
    "balance": 1250.75,
    "address": null,
    "tags": [
        "premium",
        "verified",
        "new"
    ],
    "profile": {
        "name": {
            "first": "John",
            "middle": "A.",
            "last": "Doe"
        },
        "contact": {
            "email": "john.doe@example.com",
            "phones": [
                {
                    "type": "home",
                    "number": "555-1234",
                    "primary": false
                },
                {
                    "type": "mobile",
                    "number": "555-5678",
                    "primary": true,
                    "features": [
                        "call",
                        "sms",
                        "mms"
                    ]
                }
            ]
        },
        "preferences": {
            "theme": "dark",
            "notifications": {
                "email": true,
                "sms": false,
                "push": {
                    "enabled": true,
                    "schedule": "9:00-18:00"
                }
            }
        }
    },
    "orders": [
        {
            "orderId": "ORD-1001",
            "date": "2023-05-15",
            "items": [
                {
                    "sku": "ITEM-001",
                    "quantity": 2,
                    "price": 49.99,
                    "discount": 0.1
                },
                {
                    "sku": "ITEM-005",
                    "quantity": 1,
                    "price": 129.99
                }
            ],
            "payment": {
                "method": "credit_card",
                "details": {
                    "cardType": "visa",
                    "last4": "4242"
                }
            }
        },
        {
            "orderId": "ORD-1002",
            "date": "2023-06-20",
            "items": [
                {
                    "sku": "ITEM-003",
                    "quantity": 3,
                    "price": 24.99
                }
            ],
            "payment": {
                "method": "paypal",
                "transactionId": "PAYPAL-123456"
            }
        }
    ],
    "metadata": {
        "createdAt": "2022-01-15T08:30:00Z",
        "updatedAt": "2023-07-10T14:25:00Z",
        "version": 2.1,
        "flags": {
            "hasPendingReview": false,
            "isLegacyUser": true,
            "subscription": {
                "tier": "gold",
                "expires": "2024-12-31",
                "autoRenew": true
            }
        }
    }
}`

func TestGetString(t *testing.T) {
	pathVals := map[string]interface{}{
		"/profile/contact/phones[1]/features[1]": "sms",
		"/tags[0]":                               "premium",
        "/metadata/flags/subscription/tier":      "gold",
	}

	for path, expected := range pathVals {
		v := Get(demoJson, path)
		t.Logf("%s => %s\n", path, v)
		sv, ok := v.(string)
		if !ok {
			t.Errorf("%s is not a string", path)
		}
		if sv != expected {
			t.Errorf("%s is not equal to %s", path, expected)
		}
	}

}
