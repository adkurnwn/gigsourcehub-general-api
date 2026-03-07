package usecase_member_test

import (
	"context"
	"testing"
	"time"

	usecase_member "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/member"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	gorm_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/gorm"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UnitTestSuite struct {
	suite.Suite
	usecase domain.MemberAppUsecase
}

func TestUsecaseMemberAuth(t *testing.T) {
	suite.Run(t, &UnitTestSuite{})
}

func (suite *UnitTestSuite) SetupTest() {
	gormDbRepo := new(mocks.GormRepo)

	email := "someone@mail.com"
	emailNotFound := "notfound@mail.com"
	passwordEncrypted := "$2a$10$1HSc2bCSpCUiPCrqab8JTu1QjJL4HprINKFxhylTqJs9NL233ZMiK" // password

	gormDbRepo.On("FetchOneUser", mock.Anything, gorm_model.UserFilter{}).Return(&gorm_model.User{}, nil)

	gormDbRepo.On("FetchOneUser", mock.Anything, gorm_model.UserFilter{
		Email: &email,
	}).Return(&gorm_model.User{
		Email:    email,
		Password: passwordEncrypted,
	}, nil)

	// var notfoundUser *gorm_model.User
	gormDbRepo.On("FetchOneUser", mock.Anything, gorm_model.UserFilter{
		Email: &emailNotFound,
	}).Return(nil, nil)

	// mock the newly added methods to satisfy the interface, even if unused in this test file
	gormDbRepo.On("GetDB").Return(nil)
	gormDbRepo.On("GetCVByID", mock.Anything, mock.Anything).Return(&gorm_model.CV{}, nil)
	gormDbRepo.On("UpdateCV", mock.Anything, mock.Anything).Return(nil)
	gormDbRepo.On("CreateRequest", mock.Anything, mock.Anything).Return(nil)

	suite.usecase = usecase_member.NewAppUsecase(usecase_member.RepoInjection{GormDbRepo: gormDbRepo}, time.Minute)
}

func (suite *UnitTestSuite) TestLoginNoPassword() {
	resp := suite.usecase.Login(context.TODO(), request_model.LoginRequest{
		Email:    "someone@mail.com",
		Password: "",
	})

	suite.Assert().Equal(400, resp.Status, "status should be 400")
	suite.Assert().Equal("error validation", resp.Message, "message should be error validation")
}

func (suite *UnitTestSuite) TestLoginUserNotFound() {
	resp := suite.usecase.Login(context.TODO(), request_model.LoginRequest{
		Email:    "notfound@mail.com",
		Password: "password",
	})

	suite.Assert().Equal(400, resp.Status, "status should be 400")
	suite.Assert().Equal("user not found", resp.Message, "message should be user not found")
}

func (suite *UnitTestSuite) TestLoginWrongPassword() {
	resp := suite.usecase.Login(context.TODO(), request_model.LoginRequest{
		Email:    "someone@mail.com",
		Password: "wrongpassword",
	})

	suite.Assert().Equal(400, resp.Status, "status should be 400")
	suite.Assert().Equal("Wrong password", resp.Message, "message should be Wrong password")
}

func (suite *UnitTestSuite) TestLoginSuccess() {
	resp := suite.usecase.Login(context.TODO(), request_model.LoginRequest{
		Email:    "someone@mail.com",
		Password: "password",
	})

	suite.Assert().Equal(200, resp.Status, "status should be 200")
	suite.Assert().Equal("success", resp.Message, "message should be success")
	suite.Assert().Contains(resp.Data, "token", "response should contains token")

}
