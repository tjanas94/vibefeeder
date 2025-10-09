package database

// Example usage patterns for the database client
//
// 1. Initialize the client and add middleware (in main.go):
//
//	cfg, err := config.Load()
//	if err != nil {
//		log.Fatalf("Failed to load config: %v", err)
//	}
//
//	db, err := database.New(cfg)
//	if err != nil {
//		log.Fatalf("Failed to create database client: %v", err)
//	}
//
//	e := echo.New()
//	e.Use(database.Middleware(db))
//
// 2. Use the client in handlers (retrieve from Echo context):
//
//	func GetFeeds(c echo.Context) error {
//		db := database.FromContext(c)
//		if db == nil {
//			return echo.NewHTTPError(http.StatusInternalServerError, "database not available")
//		}
//
//		var feeds []database.PublicFeedsSelect
//		err := db.DB.From("feeds").Select("*").Execute(&feeds)
//		if err != nil {
//			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
//		}
//
//		return c.JSON(http.StatusOK, feeds)
//	}
//
// 3. Query feeds (direct usage):
//
//	var feeds []database.PublicFeedsSelect
//	err = db.DB.From("feeds").Select("*").Execute(&feeds)
//	if err != nil {
//		log.Printf("Failed to fetch feeds: %v", err)
//	}
//
// Insert a new feed:
//
//	newFeed := database.PublicFeedsInsert{
//		Name:   "My Blog",
//		Url:    "https://example.com/feed.xml",
//		UserId: "user-id-here",
//	}
//	var result []database.PublicFeedsSelect
//	err = db.DB.From("feeds").Insert(newFeed).Execute(&result)
//	if err != nil {
//		log.Printf("Failed to insert feed: %v", err)
//	}
//
// Update a feed:
//
//	update := database.PublicFeedsUpdate{
//		Name: stringPtr("Updated Name"),
//	}
//	var updated []database.PublicFeedsSelect
//	err = db.DB.From("feeds").
//		Update(update).
//		Eq("id", "feed-id-here").
//		Execute(&updated)
//	if err != nil {
//		log.Printf("Failed to update feed: %v", err)
//	}
//
// Delete a feed:
//
//	var deleted []database.PublicFeedsSelect
//	err = db.DB.From("feeds").
//		Delete().
//		Eq("id", "feed-id-here").
//		Execute(&deleted)
//	if err != nil {
//		log.Printf("Failed to delete feed: %v", err)
//	}
//
// Query with filters:
//
//	var userFeeds []database.PublicFeedsSelect
//	err = db.DB.From("feeds").
//		Select("*").
//		Eq("user_id", "user-id-here").
//		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
//		Execute(&userFeeds)
//	if err != nil {
//		log.Printf("Failed to fetch user feeds: %v", err)
//	}
