// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package uaa

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/test"
	utilstest "github.com/vmware/harbor/src/common/utils/test"
	"github.com/vmware/harbor/src/common/utils/uaa"
	"github.com/vmware/harbor/src/ui/config"
)

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	server, err := utilstest.NewAdminserver(nil)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	if err := os.Setenv("ADMINSERVER_URL", server.URL); err != nil {
		panic(err)
	}
	err = config.Init()
	if err != nil {
		panic(err)
	}

	err = dao.ClearTable("project_member")
	if err != nil {
		panic(err)
	}
	err = dao.ClearTable("project_metadata")
	if err != nil {
		panic(err)
	}
	err = dao.ClearTable("access_log")
	if err != nil {
		panic(err)
	}
	err = dao.ClearTable("project")
	if err != nil {
		panic(err)
	}
	err = dao.ClearTable("harbor_user")
	if err != nil {
		panic(err)
	}

	rc := m.Run()
	os.Exit(rc)
}

func TestEnsureClient(t *testing.T) {
	assert := assert.New(t)
	auth := Auth{client: nil}
	err := auth.ensureClient()
	assert.Nil(err)
	assert.NotNil(auth.client)
}

func TestAuthenticate(t *testing.T) {
	assert := assert.New(t)
	client := &uaa.FakeClient{
		Username: "user1",
		Password: "password1",
	}
	auth := Auth{client: client}
	m1 := models.AuthModel{
		Principal: "user1",
		Password:  "password1",
	}
	u1, err1 := auth.Authenticate(m1)
	assert.Nil(err1)
	assert.NotNil(u1)
	assert.Equal("fake@fake.com", u1.Email)
	m2 := models.AuthModel{
		Principal: "wrong",
		Password:  "wrong",
	}
	u2, err2 := auth.Authenticate(m2)
	assert.NotNil(err2)
	assert.Nil(u2)
	err3 := dao.ClearTable(models.UserTable)
	assert.Nil(err3)
}

func TestOnBoardUser(t *testing.T) {
	assert := assert.New(t)
	auth := Auth{}
	um1 := &models.User{
		Username: " ",
	}
	err1 := auth.OnBoardUser(um1)
	assert.NotNil(err1)
	um2 := &models.User{
		Username: "test   ",
	}
	user2, _ := dao.GetUser(models.User{Username: "test"})
	assert.Nil(user2)
	err2 := auth.OnBoardUser(um2)
	assert.Nil(err2)
	user, _ := dao.GetUser(models.User{Username: "test"})
	assert.Equal("test", user.Realname)
	assert.Equal("test", user.Username)
	assert.Equal("test@uaa.placeholder", user.Email)
	err3 := dao.ClearTable(models.UserTable)
	assert.Nil(err3)
}

func TestPostAuthenticate(t *testing.T) {
	assert := assert.New(t)
	auth := Auth{}
	um := &models.User{
		Username: "test",
	}
	err := auth.PostAuthenticate(um)
	assert.Nil(err)
	user, _ := dao.GetUser(models.User{Username: "test"})
	assert.Equal("test@uaa.placeholder", user.Email)
	um.Email = "newEmail@new.com"
	um.Realname = "newName"
	err2 := auth.PostAuthenticate(um)
	assert.Nil(err2)
	user2, _ := dao.GetUser(models.User{Username: "test"})
	assert.Equal("newEmail@new.com", user2.Email)
	assert.Equal("newName", user2.Realname)
	err3 := dao.ClearTable(models.UserTable)
	assert.Nil(err3)

}

func TestSearchUser(t *testing.T) {
	assert := assert.New(t)
	client := &uaa.FakeClient{
		Username: "user1",
		Password: "password1",
	}
	auth := Auth{client: client}
	_, err0 := auth.SearchUser("error")
	assert.NotNil(err0)
	u1, err1 := auth.SearchUser("one")
	assert.Nil(err1)
	assert.Equal("one@email.com", u1.Email)
	_, err2 := auth.SearchUser("two")
	assert.NotNil(err2)
	user3, err3 := auth.SearchUser("none")
	assert.Nil(user3)
	assert.Nil(err3)
}
