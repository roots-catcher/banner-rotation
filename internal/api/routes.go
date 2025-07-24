package api

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")
	{
		api.POST("/banner_slot", s.addBannerToSlot)
		api.DELETE("/banner_slot", s.removeBannerFromSlot)
		api.POST("/choose_banner", s.chooseBanner)
		api.POST("/register_click", s.registerClick)
	}
}
