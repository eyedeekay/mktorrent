package mktorrent

import (
	"crypto/sha1"
	"github.com/zeebo/bencode"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const piece_len = 32768

type File struct {
	Path   []string `bencode:"path,omitempty"`
	Length int      `bencode:"length,omitempty"`
}

type InfoDict struct {
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length,omitempty"`
	Length      int    `bencode:"length,omitempty"`
	Pieces      string `bencode:"pieces,omitempty"`
	Files       []File `bencode:"files,omitempty"`
}

type Torrent struct {
	Info         InfoDict   `bencode:"info"`
	AnnounceList [][]string `bencode:"announce-list,omitempty"`
	Announce     string     `bencode:"announce,omitempty"`
	CreationDate int64      `bencode:"creation date,omitempty"`
	Comment      string     `bencode:"comment,omitempty"`
	CreatedBy    string     `bencode:"created by,omitempty"`
	UrlList      string     `bencode:"url-list,omitempty"`
}

func (t *Torrent) Save(w io.Writer) error {
	enc := bencode.NewEncoder(w)
	return enc.Encode(t)
}

func hashPiece(b []byte) []byte {
	h.Write(b)
	return h.Sum(nil)
}
func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), err
}

func PopDir(path string) []string {
	a := strings.Split(path, "/")
	_, a = a[0], a[1:]
	//  for _, b := range a {
	//    log.Println(b)
	//  }
	return a
}

var h hash.Hash

func MakeTorrent(file string, name string, url string, ann ...string) (*Torrent, error) {
	dir, err := IsDirectory(file)
	if err != nil {
		return nil, err
	}
	h = sha1.New()
	if dir {
		var t *Torrent
		if strings.HasPrefix(url, "http") || strings.HasPrefix(url, "ftp") {
			log.Println("Creating directory torrent")
			t = &Torrent{
				AnnounceList: make([][]string, 0),
				CreationDate: time.Now().Unix(),
				CreatedBy:    "mktorrent.go",
				Info: InfoDict{
					Name:        name,
					PieceLength: piece_len,
					Files:       []File{},
				},
				UrlList: url,
			}
		} else {
			t = &Torrent{
				AnnounceList: make([][]string, 0),
				CreationDate: time.Now().Unix(),
				CreatedBy:    "mktorrent.go",
				Info: InfoDict{
					Name:        name,
					PieceLength: piece_len,
					Files:       []File{},
				},
			}
		}
		// the outer list is tiers
		for _, a := range ann {
			t.AnnounceList = append(t.AnnounceList, []string{a})
		}
		i := 0
		b := make([]byte, piece_len)
		err := filepath.Walk(file,
			func(path string, info os.FileInfo, err error) error {
				if !strings.Contains(path, name+".torrent") && !strings.Contains(path, ".git"){
					if !info.IsDir() {

						log.Println("Adding File", path, "info", info.Name())
						r, err := os.Open(path)
						if err != nil {
							return err
						}

						if len(t.Info.Files) <= i {
							t.Info.Files = append(t.Info.Files, File{})
							i++
						}
						for {
							n, err := io.ReadFull(r, b)
							if err != nil && err != io.ErrUnexpectedEOF {
								return err
							}
							if err == io.ErrUnexpectedEOF {
								b = b[:n]
								t.Info.Files[i-1].Path = strings.Split(path, "/") //PopDir(path)
								t.Info.Files[i-1].Length += n
								break
							} else if n == piece_len {
								t.Info.Files[i-1].Path = strings.Split(path, "/") //PopDir(path)
								t.Info.Files[i-1].Length += n
							} else {
								//t.Info.Files[i-1].Path = strings.Split(path, "/") //PopDir(path)
								//t.Info.Files[i-1].Length += n
								//panic("short read!")
							}
						}

					}
				}
				return nil
			})
		t.Info.Pieces += string(hashPiece(b))
		if err != nil {
			return nil, err
		}
		return t, nil
	} else {
		t := &Torrent{
			AnnounceList: make([][]string, 0),
			CreationDate: time.Now().Unix(),
			CreatedBy:    "mktorrent.go",
			Info: InfoDict{
				Name:        name,
				PieceLength: piece_len,
			},
			UrlList: url,
		}

		// the outer list is tiers
		for _, a := range ann {
			t.AnnounceList = append(t.AnnounceList, []string{a})
		}

		b := make([]byte, piece_len)
		r, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		for {
			n, err := io.ReadFull(r, b)
			if err != nil && err != io.ErrUnexpectedEOF {
				return nil, err
			}
			if err == io.ErrUnexpectedEOF {
				b = b[:n]
				t.Info.Pieces += string(hashPiece(b))
				t.Info.Length += n
				break
			} else if n == piece_len {
				t.Info.Pieces += string(hashPiece(b))
				t.Info.Length += n
			} else {
				panic("short read!")
			}
		}
		return t, nil
	}

}
