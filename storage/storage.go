package storage

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"

	"main.go/lib/e"
)

type Storage interface {
	Save(p *Page) error
	PickRandom(UserName string) (*Page, error)
	Remove(p *Page) error
	IsExists(p *Page) (bool, error)
}

var ErrNoSavedPages = errors.New("no saved pages")

// Page - основной тип данных, с которы мбудет работать Storage
// Тут будет страница на каоторую ведёт ссылка, которую мы скинули боту
type Page struct {
	URL      string
	UserName string
	//Created time.Time - это для условия, что будет скидываться первая или последняя статья

}

func (p Page) Hash() (string, error) {
	h := sha1.New()

	//генерируем хэш по URL и по UserName, чтобы исключить вариант
	//в котором у нас 2 разных пользователя сохраняют одну и ту же ссылку
	if _, err := io.WriteString(h, p.URL); err != nil {
		return "", e.Wrap("can't calculate hash", err)
	}

	if _, err := io.WriteString(h, p.UserName); err != nil {
		return "", e.Wrap("can't calculate hash", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil

}
