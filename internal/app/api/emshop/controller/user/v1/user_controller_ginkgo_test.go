package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cv1 "emshop/internal/app/api/emshop/service/coupon/v1"
	gv1 "emshop/internal/app/api/emshop/service/goods/v1"
	iv1 "emshop/internal/app/api/emshop/service/inventory/v1"
	lv1 "emshop/internal/app/api/emshop/service/logistics/v1"
	ov1 "emshop/internal/app/api/emshop/service/order/v1"
	pv1 "emshop/internal/app/api/emshop/service/payment/v1"
	sv1 "emshop/internal/app/api/emshop/service/sms/v1"
	uv1 "emshop/internal/app/api/emshop/service/user/v1"
	uopv1 "emshop/internal/app/api/emshop/service/userop/v1"
	"emshop/internal/app/pkg/jwt"

	itime "emshop/pkg/common/time"

	"github.com/gin-gonic/gin"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"emshop/gin-micro/server/rest-server/validation"
)

func init() {
	gin.SetMode(gin.TestMode)
	validation.RegisterMobile(nil)
}

var _ = ginkgo.Describe("User Controller", func() {
	var (
		translator *fakeTranslator
		userSvc    *fakeUserService
		controller *userServer
	)

	ginkgo.BeforeEach(func() {
		translator = &fakeTranslator{messages: map[string]string{
			"business.captcha_required":    "captcha required",
			"business.captcha_id_required": "captcha id required",
			"business.login_failed":        "login failed",
			"business.mobile_required":     "mobile required",
			"business.password_required":   "password required",
			"business.captcha_error":       "captcha incorrect",
			"required":                     "%s is required",
			"mobile":                       "%s invalid mobile",
		}}
		userSvc = &fakeUserService{}
		controller = NewUserController(translator, &fakeServiceFactory{user: userSvc})
	})

	ginkgo.Context("Register", func() {
		ginkgo.It("returns token and profile info on success", func() {
			expectedUser := &uv1.UserDTO{
				User: uv1.User{
					ID:       100,
					NickName: "tester",
				},
				Token:     "jwt-token",
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			}
			var capturedArgs struct {
				mobile   string
				password string
				code     string
			}
			userSvc.registerFunc = func(ctx context.Context, mobile, password, code string) (*uv1.UserDTO, error) {
				capturedArgs.mobile = mobile
				capturedArgs.password = password
				capturedArgs.code = code
				return expectedUser, nil
			}

			payload := map[string]interface{}{
				"mobile":   "13800138000",
				"password": "pass@123",
				"code":     "123456",
			}
			ctx, rr := newJSONContext(http.MethodPost, "/v1/user/register", payload)

			controller.Register(ctx)

			Expect(rr.Code).To(Equal(http.StatusOK))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKeyWithValue("id", float64(expectedUser.ID)))
			Expect(resp).To(HaveKeyWithValue("nickName", expectedUser.NickName))
			Expect(resp).To(HaveKeyWithValue("token", expectedUser.Token))
			Expect(resp).To(HaveKeyWithValue("expiredAt", float64(expectedUser.ExpiresAt)))
			Expect(capturedArgs.mobile).To(Equal("13800138000"))
			Expect(capturedArgs.password).To(Equal("pass@123"))
			Expect(capturedArgs.code).To(Equal("123456"))
		})

		ginkgo.It("returns validation errors when payload is invalid", func() {
			payload := map[string]interface{}{
				"password": "pass@123",
				"code":     "123456",
			}
			ctx, rr := newJSONContext(http.MethodPost, "/v1/user/register", payload)

			controller.Register(ctx)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			resp := decodeJSON(rr)
			errors, ok := resp["error"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(errors).To(HaveKeyWithValue("Mobile", "Mobile is required"))
		})
	})

	ginkgo.Context("Login", func() {
		var originalStore captchaStore

		ginkgo.BeforeEach(func() {
			originalStore = store
		})

		ginkgo.AfterEach(func() {
			store = originalStore
		})

		ginkgo.It("authenticates user and returns profile", func() {
			expectedUser := &uv1.UserDTO{
				User: uv1.User{
					ID:       101,
					NickName: "login-user",
				},
				Token:     "token-xyz",
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			}
			userSvc.mobileLoginFunc = func(ctx context.Context, mobile, password string) (*uv1.UserDTO, error) {
				Expect(mobile).To(Equal("13800138000"))
				Expect(password).To(Equal("Pass#123"))
				return expectedUser, nil
			}

			fakeStore := &fakeCaptchaStore{
				expectedID:     "captcha-id",
				expectedAnswer: "abcde",
				verifyResult:   true,
			}
			store = fakeStore

			payload := map[string]interface{}{
				"mobile":    "13800138000",
				"password":  "Pass#123",
				"captcha":   "abcde",
				"captchaId": "captcha-id",
			}
			ctx, rr := newJSONContext(http.MethodPost, "/v1/user/pwd_login", payload)

			controller.Login(ctx)

			Expect(rr.Code).To(Equal(http.StatusOK))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKeyWithValue("id", float64(expectedUser.ID)))
			Expect(resp).To(HaveKeyWithValue("nickName", expectedUser.NickName))
			Expect(resp).To(HaveKeyWithValue("token", expectedUser.Token))
		})

		ginkgo.It("returns bad request when captcha is invalid", func() {
			store = &fakeCaptchaStore{verifyResult: false}

			payload := map[string]interface{}{
				"mobile":    "13800138000",
				"password":  "Pass#123",
				"captcha":   "wrong",
				"captchaId": "captcha-id",
			}
			ctx, rr := newJSONContext(http.MethodPost, "/v1/user/pwd_login", payload)

			controller.Login(ctx)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKey("captcha"))
		})

		ginkgo.It("validates required fields", func() {
			payload := map[string]interface{}{
				"mobile":   "13800138000",
				"password": "Pass#123",
			}
			ctx, rr := newJSONContext(http.MethodPost, "/v1/user/pwd_login", payload)

			controller.Login(ctx)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKeyWithValue("msg", "captcha required"))
		})
	})

	ginkgo.Context("Profile", func() {
		ginkgo.It("returns detail for current user", func() {
			userSvc.getFunc = func(ctx context.Context, userID uint64) (*uv1.UserDTO, error) {
				Expect(userID).To(Equal(uint64(42)))
				return &uv1.UserDTO{User: uv1.User{
					Mobile:   "13800138000",
					NickName: "Lewis",
					Birthday: itime.Time{Time: time.Date(1991, 1, 2, 0, 0, 0, 0, time.UTC)},
					Gender:   "male",
				}}, nil
			}

			ctx, rr := newJSONContext(http.MethodGet, "/v1/user/detail", nil)
			ctx.Set(jwt.KeyUserID, int(42))

			controller.GetUserDetail(ctx)

			Expect(rr.Code).To(Equal(http.StatusOK))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKeyWithValue("name", "Lewis"))
			Expect(resp).To(HaveKeyWithValue("birthday", "1991-01-02"))
		})
	})

	ginkgo.Context("Update", func() {
		ginkgo.It("updates user information with validated payload", func() {
			initialUser := &uv1.UserDTO{User: uv1.User{
				ID:       55,
				NickName: "OldName",
				Gender:   "female",
				Birthday: itime.Time{Time: time.Date(1990, 5, 10, 0, 0, 0, 0, time.UTC)},
			}}
			userSvc.getFunc = func(ctx context.Context, userID uint64) (*uv1.UserDTO, error) {
				return initialUser, nil
			}
			var updated *uv1.UserDTO
			userSvc.updateFunc = func(ctx context.Context, userDTO *uv1.UserDTO) error {
				updated = userDTO
				return nil
			}

			payload := map[string]interface{}{
				"name":     "NewName",
				"gender":   "male",
				"birthday": "1992-03-04",
			}
			ctx, rr := newJSONContext(http.MethodPatch, "/v1/user/update", payload)
			ctx.Set(jwt.KeyUserID, int(initialUser.ID))

			controller.UpdateUser(ctx)

			Expect(rr.Code).To(Equal(http.StatusOK))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKeyWithValue("Message", "用户信息更新成功"))
			Expect(updated).NotTo(BeNil())
			Expect(updated.NickName).To(Equal("NewName"))
			Expect(updated.Gender).To(Equal("male"))
			Expect(int(updated.Birthday.Unix())).To(Equal(int(time.Date(1992, 3, 4, 0, 0, 0, 0, time.Local).Unix())))
		})
	})

	ginkgo.Context("Lookup", func() {
		ginkgo.It("fetches user by mobile", func() {
			userSvc.getByMobileFunc = func(ctx context.Context, mobile string) (*uv1.UserDTO, error) {
				Expect(mobile).To(Equal("13800138000"))
				return &uv1.UserDTO{User: uv1.User{
					ID:       77,
					Mobile:   mobile,
					NickName: "Tom",
					Gender:   "male",
					Birthday: itime.Time{Time: time.Date(1995, 6, 7, 0, 0, 0, 0, time.UTC)},
				}}, nil
			}

			ctx, rr := newJSONContext(http.MethodGet, "/v1/user/get", nil)
			ctx.Request.URL.RawQuery = "mobile=13800138000"

			controller.GetByMobile(ctx)

			Expect(rr.Code).To(Equal(http.StatusOK))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKeyWithValue("mobile", "13800138000"))
			Expect(resp).To(HaveKeyWithValue("name", "Tom"))
		})

		ginkgo.It("requires mobile parameter", func() {
			ctx, rr := newJSONContext(http.MethodGet, "/v1/user/get", nil)

			controller.GetByMobile(ctx)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKeyWithValue("msg", "mobile parameter is required"))
		})

		ginkgo.It("validates id parameter when fetching by id", func() {
			ctx, rr := newJSONContext(http.MethodGet, "/v1/user/get", nil)
			ctx.Request.URL.RawQuery = "id=abc"

			controller.GetById(ctx)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKeyWithValue("msg", "invalid id parameter"))
		})

		ginkgo.It("fetches user by id", func() {
			expected := &uv1.UserDTO{User: uv1.User{
				ID:       45,
				Mobile:   "13800138000",
				NickName: "Jerry",
				Gender:   "male",
				Birthday: itime.Time{Time: time.Date(1993, 4, 5, 0, 0, 0, 0, time.UTC)},
			}}
			userSvc.getFunc = func(ctx context.Context, userID uint64) (*uv1.UserDTO, error) {
				Expect(userID).To(Equal(uint64(45)))
				return expected, nil
			}

			ctx, rr := newJSONContext(http.MethodGet, "/v1/user/get", nil)
			ctx.Request.URL.RawQuery = "id=45"

			controller.GetById(ctx)

			Expect(rr.Code).To(Equal(http.StatusOK))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKeyWithValue("id", float64(expected.ID)))
			Expect(resp).To(HaveKeyWithValue("name", expected.NickName))
		})

		ginkgo.It("returns paginated user list", func() {
			userSvc.getUserListFunc = func(ctx context.Context, pn, pSize uint32) (*uv1.UserListDTO, error) {
				Expect(pn).To(Equal(uint32(2)))
				Expect(pSize).To(Equal(uint32(5)))
				return &uv1.UserListDTO{
					TotalCount: 12,
					Items: []*uv1.UserDTO{
						{User: uv1.User{ID: 1, NickName: "A", Mobile: "1", Gender: "male", Birthday: itime.Time{Time: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)}}},
						{User: uv1.User{ID: 2, NickName: "B", Mobile: "2", Gender: "female", Birthday: itime.Time{Time: time.Date(1991, 2, 2, 0, 0, 0, 0, time.UTC)}}},
					},
				}, nil
			}

			ctx, rr := newJSONContext(http.MethodGet, "/v1/user/list", nil)
			ctx.Request.URL.RawQuery = "pn=2&pSize=5"

			controller.GetUserList(ctx)

			Expect(rr.Code).To(Equal(http.StatusOK))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKeyWithValue("total", float64(12)))
			users, ok := resp["users"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(users).To(HaveLen(2))
		})

		ginkgo.It("rejects invalid pagination parameters", func() {
			ctx, rr := newJSONContext(http.MethodGet, "/v1/user/list", nil)
			ctx.Request.URL.RawQuery = "pn=abc&pSize=5"

			controller.GetUserList(ctx)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			resp := decodeJSON(rr)
			Expect(resp).To(HaveKeyWithValue("msg", "invalid pn parameter"))
		})
	})
})

// Helpers and test fakes ----------------------------------------------------

type fakeTranslator struct {
	messages map[string]string
}

func (t *fakeTranslator) T(key string, params ...interface{}) string {
	if t == nil {
		return key
	}
	msg, ok := t.messages[key]
	if !ok {
		return key
	}
	if len(params) > 0 {
		if meta, ok := params[0].(map[string]interface{}); ok {
			field, _ := meta["Field"].(string)
			param, _ := meta["Param"].(string)
			placeholders := strings.Count(msg, "%s")
			switch placeholders {
			case 0:
				return msg
			case 1:
				return fmt.Sprintf(msg, field)
			default:
				return fmt.Sprintf(msg, field, param)
			}
		}
	}
	return msg
}

type fakeUserService struct {
	mobileLoginFunc   func(ctx context.Context, mobile, password string) (*uv1.UserDTO, error)
	registerFunc      func(ctx context.Context, mobile, password, code string) (*uv1.UserDTO, error)
	updateFunc        func(ctx context.Context, userDTO *uv1.UserDTO) error
	getFunc           func(ctx context.Context, userID uint64) (*uv1.UserDTO, error)
	getByMobileFunc   func(ctx context.Context, mobile string) (*uv1.UserDTO, error)
	getUserListFunc   func(ctx context.Context, pn, pSize uint32) (*uv1.UserListDTO, error)
	checkPasswordFunc func(ctx context.Context, password, encrypted string) (bool, error)
}

func (f *fakeUserService) MobileLogin(ctx context.Context, mobile, password string) (*uv1.UserDTO, error) {
	if f.mobileLoginFunc == nil {
		panic("mobileLoginFunc not set")
	}
	return f.mobileLoginFunc(ctx, mobile, password)
}

func (f *fakeUserService) Register(ctx context.Context, mobile, password, code string) (*uv1.UserDTO, error) {
	if f.registerFunc == nil {
		panic("registerFunc not set")
	}
	return f.registerFunc(ctx, mobile, password, code)
}

func (f *fakeUserService) Update(ctx context.Context, userDTO *uv1.UserDTO) error {
	if f.updateFunc == nil {
		panic("updateFunc not set")
	}
	return f.updateFunc(ctx, userDTO)
}

func (f *fakeUserService) Get(ctx context.Context, userID uint64) (*uv1.UserDTO, error) {
	if f.getFunc == nil {
		panic("getFunc not set")
	}
	return f.getFunc(ctx, userID)
}

func (f *fakeUserService) GetByMobile(ctx context.Context, mobile string) (*uv1.UserDTO, error) {
	if f.getByMobileFunc == nil {
		panic("getByMobileFunc not set")
	}
	return f.getByMobileFunc(ctx, mobile)
}

func (f *fakeUserService) GetUserList(ctx context.Context, pn, pSize uint32) (*uv1.UserListDTO, error) {
	if f.getUserListFunc == nil {
		panic("getUserListFunc not set")
	}
	return f.getUserListFunc(ctx, pn, pSize)
}

func (f *fakeUserService) CheckPassWord(ctx context.Context, password, encrypted string) (bool, error) {
	if f.checkPasswordFunc == nil {
		return false, fmt.Errorf("checkPasswordFunc not set")
	}
	return f.checkPasswordFunc(ctx, password, encrypted)
}

type fakeServiceFactory struct {
	user uv1.UserSrv
}

func (f *fakeServiceFactory) Goods() gv1.GoodsSrv         { return nil }
func (f *fakeServiceFactory) Users() uv1.UserSrv          { return f.user }
func (f *fakeServiceFactory) Sms() sv1.SmsSrv             { return nil }
func (f *fakeServiceFactory) Inventory() iv1.InventorySrv { return nil }
func (f *fakeServiceFactory) Order() ov1.OrderSrv         { return nil }
func (f *fakeServiceFactory) UserOp() uopv1.UserOpSrv     { return nil }
func (f *fakeServiceFactory) Coupon() cv1.CouponSrv       { return nil }
func (f *fakeServiceFactory) Payment() pv1.PaymentSrv     { return nil }
func (f *fakeServiceFactory) Logistics() lv1.LogisticsSrv { return nil }

type captchaStore interface {
	Set(id string, value string) error
	Get(id string, clear bool) string
	Verify(id, answer string, clear bool) bool
}

type fakeCaptchaStore struct {
	expectedID     string
	expectedAnswer string
	verifyResult   bool
}

func (f *fakeCaptchaStore) Set(id string, value string) error { return nil }
func (f *fakeCaptchaStore) Get(id string, clear bool) string  { return "" }
func (f *fakeCaptchaStore) Verify(id, answer string, clear bool) bool {
	if f.expectedID != "" {
		Expect(id).To(Equal(f.expectedID))
	}
	if f.expectedAnswer != "" {
		Expect(answer).To(Equal(f.expectedAnswer))
	}
	return f.verifyResult
}

func newJSONContext(method, path string, payload interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	var bodyBytes []byte
	var err error
	if payload != nil {
		bodyBytes, err = json.Marshal(payload)
		if err != nil {
			panic(err)
		}
	}

	req, err := http.NewRequest(method, path, bytes.NewReader(bodyBytes))
	if err != nil {
		panic(err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req

	return ctx, rr
}

func decodeJSON(rr *httptest.ResponseRecorder) map[string]interface{} {
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		panic(fmt.Errorf("failed to decode response: %w", err))
	}
	return resp
}

// Ensure go test sees this package-level test file.
func TestUserControllerPlaceholder(t *testing.T) {}
