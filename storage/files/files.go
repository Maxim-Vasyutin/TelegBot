package files

import (
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"main.go/lib/e"
	"main.go/storage"
)

type Storage struct {
	basePath string
}

// добавляем для всех пользователей чтение и запись
const defaultPerm = 0774

func New(basePath string) Storage {
	return Storage{basePath: basePath}

}

func (s Storage) Save(page *storage.Page) (err error) {
	//тут опеделяем способы обработки ошибок
	defer func() { err = e.WrapIfErr("can't save psge", err) }()

	//делает правильный разделитель для Windows (тут про обратный слеш)
	//тут формируем путь, куда будет сохраняться файл
	fPath := filepath.Join(s.basePath, page.UserName)

	//тут cоздаём все дериктории, переданные в путь
	if err := os.MkdirAll(fPath, defaultPerm); err != nil {
		return err
	}

	//тут формируем имя файла
	fName, err := fileName(page)
	if err != nil {
		return err
	}

	//тут дописываем имя файла к пути
	fPath = filepath.Join(fPath, fName)

	//тут создаём файл
	file, err := os.Create(fPath)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	//сериализация - привести файл к такому формату, который мы могли бы записать
	// в файл и по нему можно было бы восстановить исходеную структуру

	//Создаём энкодер и передаём ему файл, куда будет записан резульат
	// и вызываем метод энкод и передаём ему нашу страницу
	//В результате, страница будет преобразована в формат gob и записана в файл

	//тут записываем страницу в нужном формате
	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}

	return nil
}

func (s Storage) PickRandom(UserName string) (page *storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can't pick randome page", err) }()

	path := filepath.Join(s.basePath, UserName)

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	//
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(files))

	file := files[n]

	return s.decodePage(filepath.Join(path, file.Name()))
}

func (s Storage) Remove(p *storage.Page) error {
	fileName, err := fileName(p)
	if err != nil {
		return e.Wrap("can't remove file", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	if err := os.Remove(path); err != nil {
		msg := fmt.Sprintf("can't remove file %s", path)

		return e.Wrap(msg, err)
	}

	return nil
}

func (s Storage) IsExists(p *storage.Page) (bool, error) {
	fileName, err := fileName(p)
	if err != nil {
		return false, e.Wrap("can't check if file exist", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	switch _, err = os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("can't check if file %s exist", path)

		return false, e.Wrap(msg, err)
	}
	return true, nil
}

func (s Storage) decodePage(filepath string) (*storage.Page, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, e.Wrap("can't decode page", err)
	}
	defer func() { _ = f.Close() }()

	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	return &p, nil
}

func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}
