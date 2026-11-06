package handlers

import (
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/webquery"
	"net/http"
	"time"
)

func OptsFromRequest(r *http.Request) db.GetItemsOpts   { return webquery.OptsFromRequest(r) }
func TimeframeFromRequest(r *http.Request) db.Timeframe { return webquery.TimeframeFromRequest(r) }
func parseTZ(r *http.Request) *time.Location            { return webquery.Timezone(r) }
