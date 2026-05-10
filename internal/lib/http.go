package lib

type CORSConfig struct {
	AccessControlAllowOrigin   []string
	AccessControlExposeHeaders []string
	AccessControlMaxAge        int
	AccessControlAllowMethods  []string
	AccessControlAllowHeaders  []string
}
