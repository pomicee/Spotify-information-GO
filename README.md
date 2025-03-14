# Spotify API Go Wrapper

A lightweight Go wrapper for the Spotify Web API that provides simplified endpoints for retrieving song, artist, and album information.

## Features

- 🎵 Search and retrieve track information
- 👨‍🎤 Get artist details (short and full formats)
- 💿 Fetch album information
- 🔐 Automatic token management
- 🚀 Simple REST API endpoints

## Prerequisites

- Go 1.16 or higher
- Spotify Developer Account
- Spotify API Credentials (Client ID and Client Secret)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/pomicee/Spotify-information-GO
cd Spotify-information-GO
```

2. Install dependencies:
```bash
go mod init spotify-api
go mod tidy
```

## Configuration

Replace the placeholder credentials in the code with your Spotify API credentials:

```go
var (
    clientID     = "YOUR_CLIENT_ID"
    clientSecret = "YOUR_CLIENT_SECRET"
)
```

To get these credentials:
1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Create a new application
3. Copy the Client ID and Client Secret

## API Endpoints

### 1. Search for a Song
```http
GET /spotify/songs?q=SONG_NAME
```

Response:
```json
{
  "success": true,
  "track": {
    "name": "Blinding Lights",
    "fullTitle": "Blinding Lights - The Weeknd",
    "id": "0VjIjW4GlUZAMYd2vXMi3b",
    "url": "https://open.spotify.com/track/0VjIjW4GlUZAMYd2vXMi3b",
    "preview_url": "https://p.scdn.co/mp3-preview/...",
    "duration": "3:20",
    "duration_ms": 200040,
    "explicit": false,
    "popularity": 94
  }
}
```

### 2. Get Artist Information (Short)
```http
GET /spotify/artist/short?q=ARTIST_NAME
```

Response:
```json
{
  "success": true,
  "artist": {
    "name": "The Weeknd",
    "id": "1Xyo4u8uXC1ZmMpatF05PJ",
    "url": "https://open.spotify.com/artist/1Xyo4u8uXC1ZmMpatF05PJ",
    "image": "https://i.scdn.co/image/...",
    "genres": [
      "canadian contemporary r&b",
      "canadian pop",
      "pop"
    ],
    "followers": 52614183,
    "popularity": 92,
    "albums": 5,
    "singles": 43,
    "compilations": 1
  }
}
```

### 3. Get Artist Information (Full)
```http
GET /spotify/artist/full?q=ARTIST_NAME
```

Response:
```json
{
  "success": true,
  "artist": {
    "name": "The Weeknd",
    "topTracks": [
      {
        "name": "Blinding Lights",
        "popularity": 94
      }
    ],
    "albums": [
      {
        "name": "After Hours",
        "type": "album"
      }
    ],
    "albumStats": {
      "album": 5,
      "single": 43,
      "compilation": 1
    }
  }
}
```

### 4. Get Album Information
```http
GET /spotify/album?q=ALBUM_NAME
```

Response:
```json
{
  "success": true,
  "album": {
    "name": "After Hours",
    "artists": [
      {
        "name": "The Weeknd",
        "id": "1Xyo4u8uXC1ZmMpatF05PJ",
        "url": "https://open.spotify.com/artist/1Xyo4u8uXC1ZmMpatF05PJ"
      }
    ],
    "releaseDate": "2020-03-20",
    "genres": [
      "canadian contemporary r&b",
      "canadian pop"
    ],
    "totalTracks": 14,
    "popularity": 92,
    "type": "album",
    "url": "https://open.spotify.com/album/...",
    "images": [
      {
        "url": "https://i.scdn.co/image/...",
        "height": 640,
        "width": 640
      }
    ],
    "tracks": [
      {
        "name": "Blinding Lights",
        "duration": 200040,
        "trackNumber": 1,
        "url": "https://open.spotify.com/track/..."
      }
    ]
  }
}
```

## Running the Server

1. Start the server:
```bash
go run spotify.go
```

2. The server will start on port 8080:
