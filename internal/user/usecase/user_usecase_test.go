package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/user/usecase"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

// MOCK REPOSITORY (Database Palsu)
type MockUserRepository struct {
	mock.Mock
}

// meniru semua fungsi yang ada di domain.UserRepository
func (m *MockUserRepository) Create(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
}
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}
func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// MENULIS TABLE-DRIVEN TEST = fungsi GetUserByID terlebih dahulu karena alurnya paling sederhana
func TestGetUserByID(t *testing.T) {
	// Definisikan tipe struktur untuk tabel test kita
	type testCase struct {
		name           string                      // Nama skenario test
		inputID        string                      // Input yang diberikan
		mockSetup      func(m *MockUserRepository) // Cara database palsu harus merespons
		expectedError  error                       // Error yang diharapkan terjadi
		expectedResult domain.User                 // Hasil akhir yang diharapkan
	}

	// Buat daftar skenario yang mau diuji
	tests := []testCase{
		{
			name:    "Gagal karena ID kosong",
			inputID: "",
			mockSetup: func(m *MockUserRepository) {
				// Tidak memanggil database sama sekali, jadi kosongkan
			},
			expectedError:  errors.New("user ID is required"),
			expectedResult: domain.User{},
		},
		{
			name:    "Gagal karena User tidak ditemukan di Database",
			inputID: "123",
			mockSetup: func(m *MockUserRepository) {
				// Jika GetByID dipanggil dengan ID "123", pura-pura kembalikan error ErrNoDocuments
				m.On("GetByID", mock.Anything, "123").Return(domain.User{}, mongo.ErrNoDocuments).Once()
			},
			expectedError:  errors.New("user not found"),
			expectedResult: domain.User{},
		},
		{
			name:    "Sukses mengambil data User",
			inputID: "999",
			mockSetup: func(m *MockUserRepository) {
				// Jika sukses, kembalikan user dengan password yang masih terisi
				fakeUser := domain.User{
					Name:     "Mahfudz",
					Email:    "test@test.com",
					Password: "rahasia_banget", // Di usecase nanti ini harus dihilangkan
				}
				m.On("GetByID", mock.Anything, "999").Return(fakeUser, nil).Once()
			},
			expectedError: nil,
			expectedResult: domain.User{
				Name:     "Mahfudz",
				Email:    "test@test.com",
				Password: "", // Kita berharap password sudah dikosongkan oleh Usecase
			},
		},
	}

	// ==========================================
	// 3. MENJALANKAN TEST (Looping Tabel)
	// ==========================================
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Siapkan mock dan usecase
			mockRepo := new(MockUserRepository)
			tc.mockSetup(mockRepo)

			userUC := usecase.NewUserUseCase(mockRepo)

			// Eksekusi fungsinya
			result, err := userUC.GetUserByID(context.Background(), tc.inputID)

			// Validasi hasil menggunakan Testify Assert
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedResult, result)

			// Pastikan semua aturan mock yang kita buat benar-benar terpanggil
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRegister - test untuk fungsi Register
func TestRegister(t *testing.T) {
	type testCase struct {
		name          string
		inputUser     domain.User
		mockSetup     func(m *MockUserRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name: "Gagal karena email kosong",
			inputUser: domain.User{
				Email:    "",
				Password: "password123",
			},
			mockSetup: func(m *MockUserRepository) {
				// Tidak perlu call mock karena validasi email kosong
			},
			expectedError: errors.New("email is required"),
		},
		{
			name: "Gagal karena email sudah terdaftar",
			inputUser: domain.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("EmailExists", mock.Anything, "test@example.com").Return(true, nil).Once()
			},
			expectedError: errors.New("email already registered"),
		},
		{
			name: "Gagal karena password kurang dari 6 karakter",
			inputUser: domain.User{
				Email:    "test@example.com",
				Password: "12345",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("EmailExists", mock.Anything, "test@example.com").Return(false, nil).Once()
			},
			expectedError: errors.New("password must be at least 6 characters"),
		},
		{
			name: "Gagal karena error database saat EmailExists",
			inputUser: domain.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("EmailExists", mock.Anything, "test@example.com").Return(false, errors.New("connection error")).Once()
			},
			expectedError: errors.New("connection error"),
		},
		{
			name: "Gagal karena error saat Create ke database",
			inputUser: domain.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("EmailExists", mock.Anything, "test@example.com").Return(false, nil).Once()
				m.On("Create", mock.Anything, mock.MatchedBy(func(u domain.User) bool {
					return u.Email == "test@example.com"
				})).Return(errors.New("duplicate key error")).Once()
			},
			expectedError: errors.New("duplicate key error"),
		},
		{
			name: "Sukses register user baru dengan role default",
			inputUser: domain.User{
				Email:    "newuser@example.com",
				Password: "password123",
				Name:     "New User",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("EmailExists", mock.Anything, "newuser@example.com").Return(false, nil).Once()
				m.On("Create", mock.Anything, mock.MatchedBy(func(u domain.User) bool {
					return u.Email == "newuser@example.com" &&
						u.Role == domain.RoleUser &&
						u.Status == domain.StatusActive &&
						u.Password != "password123" // Password harus sudah di-hash
				})).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Sukses register dengan role custom dan nama kosong (default ke 'User')",
			inputUser: domain.User{
				Email:    "another@example.com",
				Password: "password123",
				Role:     domain.RoleAdmin,
				Name:     "",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("EmailExists", mock.Anything, "another@example.com").Return(false, nil).Once()
				m.On("Create", mock.Anything, mock.MatchedBy(func(u domain.User) bool {
					return u.Email == "another@example.com" &&
						u.Role == domain.RoleAdmin &&
						u.Name == "User"
				})).Return(nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tc.mockSetup(mockRepo)

			userUC := usecase.NewUserUseCase(mockRepo)

			err := userUC.Register(context.Background(), tc.inputUser)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestLogin - test untuk fungsi Login
func TestLogin(t *testing.T) {
	type testCase struct {
		name          string
		inputEmail    string
		inputPassword string
		mockSetup     func(m *MockUserRepository)
		expectedError error
		expectedUser  domain.User
	}

	// Password hash untuk testing (hash dari "password123")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []testCase{
		{
			name:          "Gagal karena user tidak ditemukan",
			inputEmail:    "notfound@example.com",
			inputPassword: "password123",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetByEmail", mock.Anything, "notfound@example.com").Return(domain.User{}, mongo.ErrNoDocuments).Once()
			},
			expectedError: errors.New("invalid email"),
			expectedUser:  domain.User{},
		},
		{
			name:          "Gagal karena error database saat GetByEmail",
			inputEmail:    "error@example.com",
			inputPassword: "password123",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetByEmail", mock.Anything, "error@example.com").Return(domain.User{}, errors.New("connection error")).Once()
			},
			expectedError: errors.New("connection error"),
			expectedUser:  domain.User{},
		},
		{
			name:          "Gagal karena password salah",
			inputEmail:    "test@example.com",
			inputPassword: "wrongpassword",
			mockSetup: func(m *MockUserRepository) {
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
			name:          "Gagal karena akun tidak aktif",
			inputEmail:    "inactive@example.com",
			inputPassword: "password123",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetByEmail", mock.Anything, "inactive@example.com").Return(domain.User{
					Email:    "inactive@example.com",
					Password: string(hashedPassword),
					Status:   domain.StatusInactive,
				}, nil).Once()
			},
			expectedError: errors.New("account is inactive. please contact administrator"),
			expectedUser:  domain.User{},
		},
		{
			name:          "Sukses login",
			inputEmail:    "active@example.com",
			inputPassword: "password123",
			mockSetup: func(m *MockUserRepository) {
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
				Password: "", // Password harus dikosongkan
				Status:   domain.StatusActive,
			},
		},
		{
			name:          "Sukses login dengan email uppercase (dinormalisasi)",
			inputEmail:    "ACTIVE@EXAMPLE.COM",
			inputPassword: "password123",
			mockSetup: func(m *MockUserRepository) {
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
			mockRepo := new(MockUserRepository)
			tc.mockSetup(mockRepo)

			userUC := usecase.NewUserUseCase(mockRepo)

			user, err := userUC.Login(context.Background(), tc.inputEmail, tc.inputPassword)

			assert.Equal(t, tc.expectedError, err)
			// Jangan bandingkan ID karena bson.NewObjectID() selalu generate ID baru
			// Bandingkan field lainnya saja
			assert.Equal(t, tc.expectedUser.Email, user.Email)
			assert.Equal(t, tc.expectedUser.Name, user.Name)
			assert.Equal(t, tc.expectedUser.Password, user.Password)
			assert.Equal(t, tc.expectedUser.Status, user.Status)
			assert.Equal(t, tc.expectedUser.Role, user.Role)
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetUserByIDEdgeCase - test edge case tambahan untuk GetUserByID
func TestGetUserByIDEdgeCase(t *testing.T) {
	type testCase struct {
		name           string
		inputID        string
		mockSetup      func(m *MockUserRepository)
		expectedError  error
		expectedResult domain.User
	}

	tests := []testCase{
		{
			name:    "Gagal karena error database selain ErrNoDocuments",
			inputID: "123",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, "123").Return(domain.User{}, errors.New("connection timeout")).Once()
			},
			expectedError:  errors.New("connection timeout"),
			expectedResult: domain.User{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tc.mockSetup(mockRepo)

			userUC := usecase.NewUserUseCase(mockRepo)

			result, err := userUC.GetUserByID(context.Background(), tc.inputID)

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedResult, result)
			mockRepo.AssertExpectations(t)
		})
	}
}
