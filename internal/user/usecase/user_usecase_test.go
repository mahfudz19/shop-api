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
			name:           "Gagal karena ID kosong",
			inputID:        "",
			mockSetup:      func(_ *mocks.UserRepository) {},
			expectedError:  errors.New("user ID is required"),
			expectedResult: domain.User{},
		},
		{
			name:    "Gagal karena User tidak ditemukan (404)",
			inputID: "123",
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetByID", mock.Anything, "123").Return(domain.User{}, mongo.ErrNoDocuments).Once()
			},
			expectedError:  errors.New("user not found"),
			expectedResult: domain.User{},
		},
		{
			name:    "Gagal karena Database Error",
			inputID: "123",
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetByID", mock.Anything, "123").Return(domain.User{}, errors.New("db timeout")).Once()
			},
			expectedError:  errors.New("db timeout"),
			expectedResult: domain.User{},
		},
		{
			name:    "Sukses mengambil data User",
			inputID: "999",
			mockSetup: func(m *mocks.UserRepository) {
				fakeUser := domain.User{
					Name:     "Mahfudz",
					Email:    "test@test.com",
					Password: "rahasia_banget", // Password harusnya dihapus oleh usecase
				}
				m.On("GetByID", mock.Anything, "999").Return(fakeUser, nil).Once()
			},
			expectedError: nil,
			expectedResult: domain.User{
				Name:     "Mahfudz",
				Email:    "test@test.com",
				Password: "", // Memastikan password dikosongkan
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
				Email:    "   ", // TrimSpace akan membuatnya kosong
				Password: "password123",
			},
			mockSetup:     func(_ *mocks.UserRepository) {},
			expectedError: errors.New("email is required"),
		},
		{
			name: "Gagal saat mengecek email ke DB",
			inputUser: domain.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.UserRepository) {
				m.On("EmailExists", mock.Anything, "test@example.com").Return(false, errors.New("db error")).Once()
			},
			expectedError: errors.New("db error"),
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
			name: "Gagal karena Role tidak valid",
			inputUser: domain.User{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "hacker", // Role ngasal
			},
			mockSetup: func(m *mocks.UserRepository) {
				m.On("EmailExists", mock.Anything, "test@example.com").Return(false, nil).Once()
			},
			expectedError: errors.New("role tidak valid untuk sistem"),
		},
		{
			name: "Sukses Mendaftar dengan Fallback Name 'User'",
			inputUser: domain.User{
				Email:    "noname@example.com",
				Password: "password123",
				Role:     domain.RoleUser,
				Name:     "", // Menguji fallback
			},
			mockSetup: func(m *mocks.UserRepository) {
				m.On("EmailExists", mock.Anything, "noname@example.com").Return(false, nil).Once()
				// Menggunakan MatchedBy untuk memastikan Name diubah menjadi "User"
				m.On("Create", mock.Anything, mock.MatchedBy(func(u domain.User) bool {
					return u.Name == "User"
				})).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Sukses Mendaftar Normal",
			inputUser: domain.User{
				Email:    "sukses@example.com",
				Password: "password123",
				Name:     "Mahfudz M",
				Role:     domain.RoleUser, // Wajib diisi agar lolos validasi IsValid()
			},
			mockSetup: func(m *mocks.UserRepository) {
				m.On("EmailExists", mock.Anything, "sukses@example.com").Return(false, nil).Once()
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
			name:          "Gagal karena Database Error saat GetByEmail",
			inputEmail:    "error@example.com",
			inputPassword: "password123",
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetByEmail", mock.Anything, "error@example.com").Return(domain.User{}, errors.New("db timeout")).Once()
			},
			expectedError: errors.New("db timeout"),
			expectedUser:  domain.User{},
		},
		{
			name:          "Gagal karena user tidak ditemukan (ErrNoDocuments)",
			inputEmail:    "notfound@example.com",
			inputPassword: "password123",
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetByEmail", mock.Anything, "notfound@example.com").Return(domain.User{}, mongo.ErrNoDocuments).Once()
			},
			expectedError: errors.New("invalid email or password"),
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
			expectedError: errors.New("invalid email or password"),
			expectedUser:  domain.User{},
		},
		{
			name:          "Gagal karena akun tidak aktif (Suspended/Banned)",
			inputEmail:    "banned@example.com",
			inputPassword: "password123",
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetByEmail", mock.Anything, "banned@example.com").Return(domain.User{
					Email:    "banned@example.com",
					Password: string(hashedPassword),
					Status:   domain.StatusInactive, // Status Inactive
				}, nil).Once()
			},
			expectedError: errors.New("account is inactive. please contact administrator"),
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
				Password: "", // Harus string kosong karena direset usecase
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

func TestGetAllUsers(t *testing.T) {
	type testCase struct {
		name          string
		filter        domain.UserFilter
		mockSetup     func(m *mocks.UserRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:   "Sukses GetAll Users",
			filter: domain.UserFilter{BaseQuery: domain.BaseQuery{Page: 1, Limit: 10}},
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetAll", mock.Anything, mock.Anything).
					Return(domain.UserWithPagination{
						Data: []domain.User{
							{ID: bson.NewObjectID(), Email: "user1@test.com", Name: "User 1", Role: domain.RoleUser},
							{ID: bson.NewObjectID(), Email: "user2@test.com", Name: "User 2", Role: domain.RoleAdmin},
						},
						Page:       1,
						Limit:      10,
						Total:      2,
						TotalPages: 1,
					}, nil).Once()
			},
			expectedError: nil,
		},
		{
			name:   "Gagal GetAll Users (Database Error)",
			filter: domain.UserFilter{BaseQuery: domain.BaseQuery{Page: 1, Limit: 10}},
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetAll", mock.Anything, mock.Anything).
					Return(domain.UserWithPagination{}, errors.New("database error")).Once()
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "Sukses GetAll dengan Search Filter",
			filter: domain.UserFilter{
				BaseQuery: domain.BaseQuery{
					Search: "john",
					Page:   1,
					Limit:  5,
				},
			},
			mockSetup: func(m *mocks.UserRepository) {
				m.On("GetAll", mock.Anything, mock.MatchedBy(func(f domain.UserFilter) bool {
					return f.Search == "john" && f.Page == 1 && f.Limit == 5
				})).
					Return(domain.UserWithPagination{
						Data:       []domain.User{{ID: bson.NewObjectID(), Email: "john@test.com", Name: "John Doe"}},
						Page:       1,
						Limit:      5,
						Total:      1,
						TotalPages: 1,
					}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.UserRepository)
			tc.mockSetup(mockRepo)

			userUC := usecase.NewUserUseCase(mockRepo)
			result, err := userUC.GetAllUsers(context.Background(), tc.filter)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.Data)
			}

			// Pastikan password dikosongkan
			for _, user := range result.Data {
				assert.Empty(t, user.Password)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
