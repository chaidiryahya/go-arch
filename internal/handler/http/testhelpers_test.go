package http

import "time"

// clockTestTime is a fixed instant for middleware tests so duration fields are
// deterministic and the clock.Fake seam is exercised rather than time.Now.
var clockTestTime = time.Date(2026, time.November, 26, 12, 0, 0, 0, time.UTC)
