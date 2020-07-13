package tempodb

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/grafana/tempo/tempodb/backend"
	"github.com/stretchr/testify/assert"
)

func TestTimeWindowBlockSelector(t *testing.T) {
	tests := []struct {
		name           string
		blocklist      []*backend.BlockMeta
		expected       []*backend.BlockMeta
		expectedSecond []*backend.BlockMeta
	}{
		{
			name:      "nil - nil",
			blocklist: nil,
			expected:  nil,
		},
		{
			name:      "empty - nil",
			blocklist: []*backend.BlockMeta{},
			expected:  nil,
		},
		{
			name: "two blocks returned",
			blocklist: []*backend.BlockMeta{
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				},
			},
			expected: []*backend.BlockMeta{
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				},
			},
		},
		{
			name: "three blocks choose smallest two",
			blocklist: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					TotalObjects: 0,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					TotalObjects: 1,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					TotalObjects: 0,
				},
			},
			expected: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					TotalObjects: 0,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					TotalObjects: 0,
				},
			},
		},
		{
			name: "three blocks across two windows",
			blocklist: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					TotalObjects: 0,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					TotalObjects: 1,
					EndTime:      time.Unix(1, 0),
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					TotalObjects: 0,
					EndTime:      time.Unix(1, 0),
				},
			},
			expected: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					TotalObjects: 0,
					EndTime:      time.Unix(1, 0),
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					TotalObjects: 1,
					EndTime:      time.Unix(1, 0),
				},
			},
		},
		{
			name: "two iterations of four blocks across one window",
			blocklist: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					TotalObjects: 0,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					TotalObjects: 1,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					TotalObjects: 0,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000003"),
					TotalObjects: 1,
				},
			},
			expected: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					TotalObjects: 0,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					TotalObjects: 0,
				},
			},
			expectedSecond: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					TotalObjects: 1,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000003"),
					TotalObjects: 1,
				},
			},
		},
		{
			name: "two iterations of four blocks across two windows",
			blocklist: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					TotalObjects: 0,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					TotalObjects: 1,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					TotalObjects: 0,
					EndTime:      time.Unix(1, 0),
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000003"),
					TotalObjects: 1,
					EndTime:      time.Unix(1, 0),
				},
			},
			expected: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					TotalObjects: 0,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					TotalObjects: 1,
				},
			},
			expectedSecond: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					TotalObjects: 0,
					EndTime:      time.Unix(1, 0),
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000003"),
					TotalObjects: 1,
					EndTime:      time.Unix(1, 0),
				},
			},
		},
		{
			name: "two iterations of six blocks across two windows",
			blocklist: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000004"),
					TotalObjects: 1,
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					TotalObjects: 0,
					EndTime:      time.Unix(1, 0),
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					TotalObjects: 1,
					EndTime:      time.Unix(1, 0),
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000004"),
					TotalObjects: 1,
					EndTime:      time.Unix(2, 0),
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					TotalObjects: 0,
					EndTime:      time.Unix(3, 0),
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000003"),
					TotalObjects: 1,
					EndTime:      time.Unix(3, 0),
				},
			},
			expected: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					TotalObjects: 0,
					EndTime:      time.Unix(1, 0),
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					TotalObjects: 1,
					EndTime:      time.Unix(1, 0),
				},
			},
			expectedSecond: []*backend.BlockMeta{
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					TotalObjects: 0,
					EndTime:      time.Unix(3, 0),
				},
				{
					BlockID:      uuid.MustParse("00000000-0000-0000-0000-000000000003"),
					TotalObjects: 1,
					EndTime:      time.Unix(3, 0),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := newTimeWindowBlockSelector(tt.blocklist, time.Second, 100)

			actual, _ := selector.BlocksToCompact()
			assert.Equal(t, tt.expected, actual)

			actual, _ = selector.BlocksToCompact()
			assert.Equal(t, tt.expectedSecond, actual)
		})
	}
}

func TestTimeWindowBlockSelectorActiveWindow(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		blocklist      []*backend.BlockMeta
		expected       []*backend.BlockMeta
		expectedSecond []*backend.BlockMeta
	}{
		{
			name:      "nil - nil",
			blocklist: nil,
			expected:  nil,
		},
		{
			name:      "empty - nil",
			blocklist: []*backend.BlockMeta{},
			expected:  nil,
		},
		{
			name: "two blocks returned",
			blocklist: []*backend.BlockMeta{
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					EndTime: now,
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					EndTime: now,
				},
			},
			expected: []*backend.BlockMeta{
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					EndTime: now,
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					EndTime: now,
				},
			},
		},
		{
			name: "three blocks choose smallest two",
			blocklist: []*backend.BlockMeta{
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					CompactionLevel: 0,
					EndTime:         now,
				},
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					CompactionLevel: 1,
					EndTime:         now,
				},
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					CompactionLevel: 0,
					EndTime:         now,
				},
			},
			expected: []*backend.BlockMeta{
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					CompactionLevel: 0,
					EndTime:         now,
				},
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					CompactionLevel: 0,
					EndTime:         now,
				},
			},
		},
		{
			name: "three blocks choose larger two",
			blocklist: []*backend.BlockMeta{
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					CompactionLevel: 1,
					EndTime:         now,
				},
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					CompactionLevel: 0,
					EndTime:         now,
				},
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					CompactionLevel: 1,
					EndTime:         now,
				},
			},
			expected: []*backend.BlockMeta{
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					CompactionLevel: 1,
					EndTime:         now,
				},
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					CompactionLevel: 1,
					EndTime:         now,
				},
			},
		},
		{
			name: "three blocks choose none",
			blocklist: []*backend.BlockMeta{
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					CompactionLevel: 0,
					EndTime:         now,
				},
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					CompactionLevel: 1,
					EndTime:         now,
				},
				{
					BlockID:         uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					CompactionLevel: 2,
					EndTime:         now,
				},
			},
			expected: nil,
		},
		{
			name: "four blocks across two time windows",
			blocklist: []*backend.BlockMeta{
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					EndTime: now,
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					EndTime: now,
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					EndTime: now.Add(-24 * time.Hour),
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000003"),
					EndTime: now.Add(-24 * time.Hour),
				},
			},
			expected: []*backend.BlockMeta{
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					EndTime: now,
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					EndTime: now,
				},
			},
		},
		{
			name: "four blocks across two time windows.  skip buffer",
			blocklist: []*backend.BlockMeta{
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					EndTime: now,
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					EndTime: now,
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					EndTime: now.Add(-24 * time.Hour),
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000003"),
					EndTime: now.Add(-24 * time.Hour),
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000004"),
					EndTime: now.Add(-48 * time.Hour),
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000005"),
					EndTime: now.Add(-48 * time.Hour),
				},
			},
			expected: []*backend.BlockMeta{
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					EndTime: now,
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					EndTime: now,
				},
			},
			expectedSecond: []*backend.BlockMeta{
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000004"),
					EndTime: now.Add(-48 * time.Hour),
				},
				{
					BlockID: uuid.MustParse("00000000-0000-0000-0000-000000000005"),
					EndTime: now.Add(-48 * time.Hour),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := newTimeWindowBlockSelector(tt.blocklist, 24*time.Hour, 100)

			actual, _ := selector.BlocksToCompact()
			assert.Equal(t, tt.expected, actual)

			actual, _ = selector.BlocksToCompact()
			assert.Equal(t, tt.expectedSecond, actual)
		})
	}
}