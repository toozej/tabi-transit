package sqlcgen

import "testing"

func TestGeneratedEnumsScanAndValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		scan func() (string, error)
	}{
		{"notification delivery", func() (string, error) {
			var value AppNotificationDeliveryStatus
			err := value.Scan([]byte("sent"))
			return string(value), err
		}},
		{"subscription", func() (string, error) {
			var value AppSubscriptionType
			err := value.Scan("service_alert")
			return string(value), err
		}},
		{"feed version", func() (string, error) {
			var value CatalogFeedVersionStatus
			err := value.Scan("active")
			return string(value), err
		}},
		{"freshness", func() (string, error) {
			var value RealtimeFreshnessStatus
			err := value.Scan("fresh")
			return string(value), err
		}},
		{"mode", func() (string, error) { var value TransitMode; err := value.Scan("bus"); return string(value), err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.scan()
			if err != nil || got == "" {
				t.Fatalf("scan = %q, %v", got, err)
			}
		})
	}

	null := NullTransitMode{}
	if err := null.Scan(nil); err != nil || null.Valid {
		t.Fatalf("null scan = %#v, %v", null, err)
	}
	value, err := null.Value()
	if err != nil || value != nil {
		t.Fatalf("null value = %#v, %v", value, err)
	}
}

func TestGeneratedQueriesSatisfyContract(t *testing.T) {
	t.Parallel()
	queries := New(nil)
	if queries == nil {
		t.Fatal("New returned nil")
	}
	var _ Querier = queries
}
