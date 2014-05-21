package models

import (
  "fmt"
  "log"

  "github.com/astaxie/beego"
  "github.com/dockboard/docker-registry/utils"
  _ "github.com/go-sql-driver/mysql"
  "github.com/go-xorm/xorm"
)

var x *xorm.Engine

type User struct {
  Id       int64
  Username string `xorm:"VARCHAR(255)"`
  Password string `xorm:"VARCHAR(255)"`
  Email    string `xorm:"VARCHAR(255)"`
  Token    string `xorm:"VARCHAR(255)"` //MD5(Username+Password+Timestamp)
}

type Image struct {
  Id         int64
  ImageId    string `xorm:"VARCHAR(255)"`
  JSON       string `xorm:"TEXT"`
  ParentJSON string `xorm:"TEXT"`
  Checksum   string `xorm:"TEXT"`
  Payload    string `xorm:"TEXT"`
  Uploaded   bool   `xorm:"Bool"`
  CheckSumed bool   `xorm:"Bool"`
}

type Repository struct {
  Id         int64
  Namespace  string `xorm:"VARCHAR(255)"`
  Repository string `xorm:"VARCHAR(255)"`
  TagName    string `xorm:"VARCHAR(255)"`
  TagJSON    string `xorm:"TEXT"`
  Tag        string `xorm:"VARCHAR(255)"`
}

func setEngine() {
  host   := utils.Cfg.MustValue("mysql", "host")
  name   := utils.Cfg.MustValue("mysql", "name")
  user   := utils.Cfg.MustValue("mysql", "user")
  passwd := utils.Cfg.MustValue("mysql", "passwd")

  var err error
  conn := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8", user, passwd, host, name)
  beego.Trace("Initialized database connStr ->", conn)

  x, err = xorm.NewEngine("mysql", conn)
  if err != nil {
    log.Fatalf("models.init -> fail to conntect database: %v", err)
  }

  x.ShowDebug = true
  x.ShowErr = true
  x.ShowSQL = true

  beego.Trace("Initialized database ->", dbName)

}

// InitDb initializes the database.
func InitDb() {
  setEngine()
  err := x.Sync(new(User), new(Image), new(Repository))
  if err != nil {
    log.Fatalf("models.init -> fail to sync database: %v", err)
  }
}

func GetImageById(imageId string) (returnImage *Image, err error) {
  returnImage = new(Image)
  rows, err := x.Where("image_id=?", imageId).Rows(returnImage)
  defer rows.Close()
  if err != nil {
    returnImage = nil
    return
  }
  if rows.Next() {
    rows.Scan(returnImage)
  } else {
    returnImage = nil
  }

  return
}

type AuthError string

func (e AuthError) Error() string {
  return string(e)
}

type OrmError string

func (e OrmError) Error() string {
  return string(e)
}

func GetRegistryUserByUserName(mUserName string) (returnRegistryUser *User, err error) {
  returnRegistryUser = new(User)
  rows, err := x.Where("username=?", mUserName).Rows(returnRegistryUser)
  if rows.Next() {
    rows.Scan(returnRegistryUser)
    return returnRegistryUser, nil
  } else {
    return nil, OrmError("get user by name error")
  }

}

func GetRegistryUserByToken(mUserName string, mToken string) (returnRegistryUser *User, err error) {
  returnRegistryUser = new(User)
  rows, err := x.Where("username=? and token=?", mUserName, mToken).Rows(returnRegistryUser)
  if rows.Next() {
    rows.Scan(returnRegistryUser)
    return returnRegistryUser, nil
  } else {
    return nil, OrmError("get user by token error")
  }

}

func UpRegistryUser(upRegistryUser *User) (err error) {
  _, err = x.Id(upRegistryUser.Id).Update(upRegistryUser)
  if err != nil {
    return err
  } else {
    return nil
  }
}

func GetRegistryUserAuth(authUsername string, authPassword string) (err error) {
  mRegistryUser := new(User)
  rows, err := x.Where("username=? and password=?", authUsername, authPassword).Rows(mRegistryUser)

  if rows.Next() {
    return nil
  } else {
    return AuthError("Auth Error")
  }
}

func InsertOneImage(putRegistryImage *Image) (affected int64, err error) {
  affected, err = x.InsertOne(putRegistryImage)
  return
}

func UpOneImage(putRegistryImage *Image) (affected int64, err error) {
  affected, err = x.Id(putRegistryImage.Id).Update(putRegistryImage)
  fmt.Println("putRegistryImage.ImageCheckSumed:", putRegistryImage.CheckSumed, "___affected:", affected, "___err:", err)
  return
}

func InsertOneTag(insertRegistryRepositorieTag *Repository) (affected int64, err error) {
  affected, err = x.InsertOne(insertRegistryRepositorieTag)
  return
}

func UpOneTag(upRegistryRepositorieTag *Repository) (affected int64, err error) {
  affected, err = x.Id(upRegistryRepositorieTag.Id).Update(upRegistryRepositorieTag)
  return
}

func PutOneTag(upRegistryRepositorieTag *Repository) (affected int64, err error) {
  rows, err := x.Where("repositorie_tag_name=? and repositorie_tag_namespace=? and repositorie_tag_repository=?",
    upRegistryRepositorieTag.TagName,
    upRegistryRepositorieTag.Namespace,
    upRegistryRepositorieTag.Repository).Rows(upRegistryRepositorieTag)
  defer rows.Close()
  if rows.Next() {
    x.Id(upRegistryRepositorieTag.Id).Delete(upRegistryRepositorieTag)
  }
  affected, err = x.InsertOne(upRegistryRepositorieTag)
  return
}