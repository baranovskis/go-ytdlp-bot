package ytdlp

// Info youtube-dl info
type Info struct {
	// Generated from youtube-dl README using:
	// sed -e 's/ - `\(.*\)` (\(.*\)): \(.*\)/\1 \2 `json:"\1"` \/\/ \3/' | sed -e 's/numeric/float64/' | sed -e 's/boolean/bool/' | sed -e 's/_id/ID/'  | sed -e 's/_count/Count/'| sed -e 's/_uploader/Uploader/' | sed -e 's/_key/Key/' | sed -e 's/_year/Year/' | sed -e 's/_title/Title/' | sed -e 's/_rating/Rating/'  | sed -e 's/_number/Number/'  | awk '{print toupper(substr($0, 0, 1))  substr($0, 2)}'
	ID                 string  `json:"id"`                   // Video identifier
	Title              string  `json:"title"`                // Video title
	URL                string  `json:"url"`                  // Video URL
	AltTitle           string  `json:"alt_title"`            // A secondary title of the video
	DisplayID          string  `json:"display_id"`           // An alternative identifier for the video
	Uploader           string  `json:"uploader"`             // Full name of the video uploader
	License            string  `json:"license"`              // License name the video is licensed under
	Creator            string  `json:"creator"`              // The creator of the video
	ReleaseDate        string  `json:"release_date"`         // The date (YYYYMMDD) when the video was released
	Timestamp          float64 `json:"timestamp"`            // UNIX timestamp of the moment the video became available
	UploadDate         string  `json:"upload_date"`          // Video upload date (YYYYMMDD)
	UploaderID         string  `json:"uploader_id"`          // Nickname or id of the video uploader
	Channel            string  `json:"channel"`              // Full name of the channel the video is uploaded on
	ChannelID          string  `json:"channel_id"`           // Id of the channel
	Location           string  `json:"location"`             // Physical location where the video was filmed
	Duration           float64 `json:"duration"`             // Length of the video in seconds
	ViewCount          float64 `json:"view_count"`           // How many users have watched the video on the platform
	LikeCount          float64 `json:"like_count"`           // Number of positive ratings of the video
	DislikeCount       float64 `json:"dislike_count"`        // Number of negative ratings of the video
	RepostCount        float64 `json:"repost_count"`         // Number of reposts of the video
	AverageRating      float64 `json:"average_rating"`       // Average rating give by users, the scale used depends on the webpage
	CommentCount       float64 `json:"comment_count"`        // Number of comments on the video
	AgeLimit           float64 `json:"age_limit"`            // Age restriction for the video (years)
	IsLive             bool    `json:"is_live"`              // Whether this video is a live stream or a fixed-length video
	StartTime          float64 `json:"start_time"`           // Time in seconds where the reproduction should start, as specified in the URL
	EndTime            float64 `json:"end_time"`             // Time in seconds where the reproduction should end, as specified in the URL
	Extractor          string  `json:"extractor"`            // Name of the extractor
	ExtractorKey       string  `json:"extractor_key"`        // Key name of the extractor
	Epoch              float64 `json:"epoch"`                // Unix epoch when creating the file
	Autonumber         float64 `json:"autonumber"`           // Five-digit number that will be increased with each download, starting at zero
	Playlist           string  `json:"playlist"`             // Name or id of the playlist that contains the video
	PlaylistIndex      float64 `json:"playlist_index"`       // Index of the video in the playlist padded with leading zeros according to the total length of the playlist
	PlaylistID         string  `json:"playlist_id"`          // Playlist identifier
	PlaylistTitle      string  `json:"playlist_title"`       // Playlist title
	PlaylistUploader   string  `json:"playlist_uploader"`    // Full name of the playlist uploader
	PlaylistUploaderID string  `json:"playlist_uploader_id"` // Nickname or id of the playlist uploader

	// Available for the video that belongs to some logical chapter or section:
	Chapter       string  `json:"chapter"`        // Name or title of the chapter the video belongs to
	ChapterNumber float64 `json:"chapter_number"` // Number of the chapter the video belongs to
	ChapterID     string  `json:"chapter_id"`     // Id of the chapter the video belongs to

	// Available for the video that is an episode of some series or programme:
	Series        string  `json:"series"`         // Title of the series or programme the video episode belongs to
	Season        string  `json:"season"`         // Title of the season the video episode belongs to
	SeasonNumber  float64 `json:"season_number"`  // Number of the season the video episode belongs to
	SeasonID      string  `json:"season_id"`      // Id of the season the video episode belongs to
	Episode       string  `json:"episode"`        // Title of the video episode
	EpisodeNumber float64 `json:"episode_number"` // Number of the video episode within a season
	EpisodeID     string  `json:"episode_id"`     // Id of the video episode

	// Available for the media that is a track or a part of a music album:
	Track       string  `json:"track"`        // Title of the track
	TrackNumber float64 `json:"track_number"` // Number of the track within an album or a disc
	TrackID     string  `json:"track_id"`     // Id of the track
	Artist      string  `json:"artist"`       // Artist(s) of the track
	Genre       string  `json:"genre"`        // Genre(s) of the track
	Album       string  `json:"album"`        // Title of the album the track belongs to
	AlbumType   string  `json:"album_type"`   // Type of the album
	AlbumArtist string  `json:"album_artist"` // List of all artists appeared on the album
	DiscNumber  float64 `json:"disc_number"`  // Number of the disc or other physical medium the track belongs to
	ReleaseYear float64 `json:"release_year"` // Year (YYYY) when the album was released

	Type        string `json:"_type"`
	Direct      bool   `json:"direct"`
	WebpageURL  string `json:"webpage_url"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	// don't unmarshal, populated from image thumbnail file
	ThumbnailBytes []byte      `json:"-"`
	Thumbnails     []Thumbnail `json:"thumbnails"`

	Formats   []Format              `json:"formats"`
	Subtitles map[string][]Subtitle `json:"subtitles"`

	// Playlist entries if _type is playlist
	Entries []Info `json:"entries"`

	// Info can also be a mix of Info and one Format
	Format
}

type Thumbnail struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	Preference int    `json:"preference"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Resolution string `json:"resolution"`
}

// Format youtube-dl downloadable format
type Format struct {
	Ext            string            `json:"ext"`             // Video filename extension
	Format         string            `json:"format"`          // A human-readable description of the format
	FormatID       string            `json:"format_id"`       // Format code specified by `--format`
	FormatNote     string            `json:"format_note"`     // Additional info about the format
	Width          float64           `json:"width"`           // Width of the video
	Height         float64           `json:"height"`          // Height of the video
	Resolution     string            `json:"resolution"`      // Textual description of width and height
	TBR            float64           `json:"tbr"`             // Average bitrate of audio and video in KBit/s
	ABR            float64           `json:"abr"`             // Average audio bitrate in KBit/s
	ACodec         string            `json:"acodec"`          // Name of the audio codec in use
	ASR            float64           `json:"asr"`             // Audio sampling rate in Hertz
	VBR            float64           `json:"vbr"`             // Average video bitrate in KBit/s
	FPS            float64           `json:"fps"`             // Frame rate
	VCodec         string            `json:"vcodec"`          // Name of the video codec in use
	Container      string            `json:"container"`       // Name of the container format
	Filename       string            `json:"filename"`        //
	Filesize       float64           `json:"filesize"`        // The number of bytes, if known in advance
	FilesizeApprox float64           `json:"filesize_approx"` // An estimate for the number of bytes
	Protocol       string            `json:"protocol"`        // The protocol that will be used for the actual download
	HTTPHeaders    map[string]string `json:"http_headers"`
}

// Subtitle youtube-dl subtitle
type Subtitle struct {
	URL      string `json:"url"`
	Ext      string `json:"ext"`
	Language string `json:"-"`
	// don't unmarshal, populated from subtitle file
	Bytes []byte `json:"-"`
}
