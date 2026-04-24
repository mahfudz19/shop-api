package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/mocks"
	"github.com/username/shop-api/internal/user/usecase"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

func TestGetUserByID(t *testing.T) {
	type testCase struct {
		name           string
		inputID        string
		mockSetup      func(m *mocks.UserRepository)
		expectedError  error
		expectedResult domain.User
	}

	tests := []testCase{
		{
			name:    "Gagal karena ID kosong",
			inputID: "",
			mockSetup: func(_ *mocks.UserRepository) {
			},
			expectedError:  errors.New("user ID is required"),
			expectedResult: domain.User{},
		},
		{
			name:    "Gagal karena User tidak ditemukan di Database",
			inputID: "123",
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetByID", mock.Anything, "123").Return(domain.User{}, mongo.ErrNoDocuments).Once()
			},
			expectedError:  errors.New("user not found"),
			expectedResult: domain.User{},
		},
		{
			name:    "Sukses mengambil data User",
			inputID: "999",
			mockSetup: func(m *mocks.UserRepository) {
				fakeUser := domain.User{
					Name:     "Mahfudz",
					Email:    "test@test.com",
					Password: "rahasia_banget",
				}
				m.On("GetByID", mock.Anything, "999").Return(fakeUser, nil).Once()
			},
			expectedError: nil,
			expectedResult: domain.User{
				Name:     "Mahfudz",
				Email:    "test@test.com",
				Password: "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.UserRepository)
			tc.mockSetup(mockRepo)

			userUC := usecase.NewUserUseCase(mockRepo)
			result, err := userUC.GetUserByID(context.Background(), tc.inputID)

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedResult, result)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRegister(t *testing.T) {
	type testCase struct {
		name          string
		inputUser     domain.User
		mockSetup     func(m *mocks.UserRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name: "Gagal karena email kosong",
			inputUser: domain.User{
				Email:    "",
				Password: "password123",
			},
			mockSetup:     func(_ *mocks.UserRepository) {},
			expectedError: errors.New("email is required"),
		},
		{
			name: "Gagal karena password kurang dari 6 karakter",
			inputUser: domain.User{
				Email:    "test@example.com",
				Password: "123",
			},
			mockSetup: func(m *mocks.UserRepository) {
				m.On("EmailExists", mock.Anything, "test@example.com").Return(false, nil).Once()
			},
			expectedError: errors.New("password must be at least 6 characters"),
		},
		{
			name: "Gagal karena email sudah terdaftar",
			inputUser: domain.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.UserRepository) {
				m.On("EmailExists", mock.Anything, "test@example.com").Return(true, nil).Once()
			},
			expectedError: errors.New("email already registered"),
		},
		{
			name: "Sukses Mendaftar",
			inputUser: domain.User{
				Email:    "sukses@example.com",
				Password: "password123",
				Name:     "Mahfudz M",
			},
			mockSetup: func(m *mocks.UserRepository) {
				// 1. Pastikan email belum ada
				m.On("EmailExists", mock.Anything, "sukses@example.com").Return(false, nil).Once()
				// 2. Pastikan fungsi Create dipanggil, kembalikan nilai sukses (nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.User")).Return(nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.UserRepository)
			tc.mockSetup(mockRepo)

			userUC := usecase.NewUserUseCase(mockRepo)
			err := userUC.Register(context.Background(), tc.inputUser)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLogin(t *testing.T) {
	type testCase struct {
		name          string
		inputEmail    string
		inputPassword string
		mockSetup     func(m *mocks.UserRepository)
		expectedError error
		expectedUser  domain.User
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []testCase{
		{
			name:          "Gagal karena user tidak ditemukan",
			inputEmail:    "notfound@example.com",
			inputPassword: "password123",
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetByEmail", mock.Anything, "notfound@example.com").Return(domain.User{}, mongo.ErrNoDocuments).Once()
			},
			expectedError: errors.New("invalid email"),
			expectedUser:  domain.User{},
		},
		{
			name:          "Gagal karena password salah",
			inputEmail:    "test@example.com",
			inputPassword: "wrongpassword",
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetByEmail", mock.Anything, "test@example.com").Return(domain.User{
					Email:    "test@example.com",
					Password: string(hashedPassword),
					Status:   domain.StatusActive,
				}, nil).Once()
			},
			expectedError: errors.New("invalid password"),
			expectedUser:  domain.User{},
		},
		{
			name:          "Sukses login",
			inputEmail:    "active@example.com",
			inputPassword: "password123",
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetByEmail", mock.Anything, "active@example.com").Return(domain.User{
					ID:       bson.NewObjectID(),
					Email:    "active@example.com",
					Name:     "Active User",
					Password: string(hashedPassword),
					Status:   domain.StatusActive,
				}, nil).Once()
			},
			expectedError: nil,
			expectedUser: domain.User{
				Email:    "active@example.com",
				Name:     "Active User",
				Password: "",
				Status:   domain.StatusActive,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.UserRepository)
			tc.mockSetup(mockRepo)

			userUC := usecase.NewUserUseCase(mockRepo)
			user, err := userUC.Login(context.Background(), tc.inputEmail, tc.inputPassword)

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedUser.Email, user.Email)
			assert.Equal(t, tc.expectedUser.Name, user.Name)
			assert.Equal(t, tc.expectedUser.Password, user.Password)
			mockRepo.AssertExpectations(t)
		})
	}
}
