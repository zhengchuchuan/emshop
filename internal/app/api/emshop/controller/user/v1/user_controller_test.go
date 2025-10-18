package user

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
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
    "github.com/stretchr/testify/assert"

    "emshop/gin-micro/server/rest-server/validation"
)

func init() {
    gin.SetMode(gin.TestMode)
    validation.RegisterMobile(nil)
}

func TestUserController_Register(t *testing.T) {
    translator := &fakeTranslator{messages: map[string]string{
        "business.captcha_required":    "captcha required",
        "business.captcha_id_required": "captcha id required",
        "business.login_failed":        "login failed",
        "business.mobile_required":     "mobile required",
        "business.password_required":   "password required",
        "business.captcha_error":       "captcha incorrect",
        "required":                     "%s is required",
        "mobile":                       "%s invalid mobile",
    }}
    userSvc := &fakeUserService{}
    controller := NewUserController(translator, &fakeServiceFactory{user: userSvc})

    t.Run("success", func(t *testing.T) {
        expected := &uv1.UserDTO{User: uv1.User{ID: 100, NickName: "tester"}, Token: "jwt-token", ExpiresAt: time.Now().Add(time.Hour).Unix()}
        var captured struct{ mobile, password, code string }
        userSvc.registerFunc = func(ctx context.Context, mobile, password, code string) (*uv1.UserDTO, error) {
            captured.mobile, captured.password, captured.code = mobile, password, code
            return expected, nil
        }

        payload := map[string]any{"mobile": "13800138000", "password": "pass@123", "code": "123456"}
        ctx, rr := newJSONContext(http.MethodPost, "/v1/user/register", payload)
        controller.Register(ctx)

        assert.Equal(t, http.StatusOK, rr.Code)
        resp := decodeJSON(rr)
        assert.Equal(t, float64(expected.ID), resp["id"])
        assert.Equal(t, expected.NickName, resp["nickName"])
        assert.Equal(t, expected.Token, resp["token"])
        assert.Equal(t, float64(expected.ExpiresAt), resp["expiredAt"])
        assert.Equal(t, "13800138000", captured.mobile)
        assert.Equal(t, "pass@123", captured.password)
        assert.Equal(t, "123456", captured.code)
    })

    t.Run("invalid payload", func(t *testing.T) {
        payload := map[string]any{"password": "pass@123", "code": "123456"}
        ctx, rr := newJSONContext(http.MethodPost, "/v1/user/register", payload)
        controller.Register(ctx)
        assert.Equal(t, http.StatusBadRequest, rr.Code)
        resp := decodeJSON(rr)
        errs := resp["error"].(map[string]any)
        _, ok := errs["Mobile"]
        assert.True(t, ok)
    })
}

func TestUserController_Login(t *testing.T) {
    translator := &fakeTranslator{messages: map[string]string{
        "business.captcha_required":    "captcha required",
        "business.captcha_id_required": "captcha id required",
        "business.login_failed":        "login failed",
        "business.captcha_error":       "captcha incorrect",
    }}
    userSvc := &fakeUserService{}
    controller := NewUserController(translator, &fakeServiceFactory{user: userSvc})

    t.Run("success", func(t *testing.T) {
        expected := &uv1.UserDTO{User: uv1.User{ID: 101, NickName: "login-user"}, Token: "token-xyz", ExpiresAt: time.Now().Add(time.Hour).Unix()}
        userSvc.mobileLoginFunc = func(ctx context.Context, mobile, password string) (*uv1.UserDTO, error) {
            assert.Equal(t, "13800138000", mobile)
            assert.Equal(t, "Pass#123", password)
            return expected, nil
        }
        // fake captcha store
        original := store
        defer func() { store = original }()
        store = &fakeCaptchaStore{expectedID: "captcha-id", expectedAnswer: "abcde", verifyResult: true}

        payload := map[string]any{"mobile": "13800138000", "password": "Pass#123", "captcha": "abcde", "captchaId": "captcha-id"}
        ctx, rr := newJSONContext(http.MethodPost, "/v1/user/pwd_login", payload)
        controller.Login(ctx)
        assert.Equal(t, http.StatusOK, rr.Code)
        resp := decodeJSON(rr)
        assert.Equal(t, float64(expected.ID), resp["id"])
        assert.Equal(t, expected.NickName, resp["nickName"])
        assert.Equal(t, expected.Token, resp["token"])
    })

    t.Run("invalid captcha", func(t *testing.T) {
        original := store
        defer func() { store = original }()
        store = &fakeCaptchaStore{verifyResult: false}
        payload := map[string]any{"mobile": "13800138000", "password": "Pass#123", "captcha": "wrong", "captchaId": "captcha-id"}
        ctx, rr := newJSONContext(http.MethodPost, "/v1/user/pwd_login", payload)
        controller.Login(ctx)
        assert.Equal(t, http.StatusBadRequest, rr.Code)
        resp := decodeJSON(rr)
        _, ok := resp["captcha"]
        assert.True(t, ok)
    })

    t.Run("missing captcha", func(t *testing.T) {
        payload := map[string]any{"mobile": "13800138000", "password": "Pass#123"}
        ctx, rr := newJSONContext(http.MethodPost, "/v1/user/pwd_login", payload)
        controller.Login(ctx)
        assert.Equal(t, http.StatusBadRequest, rr.Code)
        resp := decodeJSON(rr)
        assert.Equal(t, "captcha required", resp["msg"])
    })
}

func TestUserController_Profile_Update_Lookup(t *testing.T) {
    translator := &fakeTranslator{}
    userSvc := &fakeUserService{}
    controller := NewUserController(translator, &fakeServiceFactory{user: userSvc})

    t.Run("profile detail", func(t *testing.T) {
        userSvc.getFunc = func(ctx context.Context, userID uint64) (*uv1.UserDTO, error) {
            assert.Equal(t, uint64(42), userID)
            return &uv1.UserDTO{User: uv1.User{Mobile: "13800138000", NickName: "Lewis", Birthday: itime.Time{Time: time.Date(1991, 1, 2, 0, 0, 0, 0, time.UTC)}, Gender: "male"}}, nil
        }
        ctx, rr := newJSONContext(http.MethodGet, "/v1/user/detail", nil)
        ctx.Set(jwt.KeyUserID, int(42))
        controller.GetUserDetail(ctx)
        assert.Equal(t, http.StatusOK, rr.Code)
        resp := decodeJSON(rr)
        assert.Equal(t, "Lewis", resp["name"])
        assert.Equal(t, "1991-01-02", resp["birthday"])
    })

    t.Run("update", func(t *testing.T) {
        initial := &uv1.UserDTO{User: uv1.User{ID: 55, NickName: "OldName", Gender: "female", Birthday: itime.Time{Time: time.Date(1990, 5, 10, 0, 0, 0, 0, time.UTC)}}}
        userSvc.getFunc = func(ctx context.Context, userID uint64) (*uv1.UserDTO, error) { return initial, nil }
        var updated *uv1.UserDTO
        userSvc.updateFunc = func(ctx context.Context, u *uv1.UserDTO) error { updated = u; return nil }

        payload := map[string]any{"name": "NewName", "gender": "male", "birthday": "1992-03-04"}
        ctx, rr := newJSONContext(http.MethodPatch, "/v1/user/update", payload)
        ctx.Set(jwt.KeyUserID, int(initial.ID))
        controller.UpdateUser(ctx)
        assert.Equal(t, http.StatusOK, rr.Code)
        resp := decodeJSON(rr)
        assert.Equal(t, "用户信息更新成功", resp["Message"])
        if assert.NotNil(t, updated) {
            assert.Equal(t, "NewName", updated.NickName)
            assert.Equal(t, "male", updated.Gender)
        }
    })

    t.Run("lookup by mobile", func(t *testing.T) {
        userSvc.getByMobileFunc = func(ctx context.Context, mobile string) (*uv1.UserDTO, error) {
            assert.Equal(t, "13800138000", mobile)
            return &uv1.UserDTO{User: uv1.User{ID: 77, Mobile: mobile, NickName: "Tom", Gender: "male", Birthday: itime.Time{Time: time.Date(1995, 6, 7, 0, 0, 0, 0, time.UTC)}}}, nil
        }
        ctx, rr := newJSONContext(http.MethodGet, "/v1/user/get", nil)
        ctx.Request.URL.RawQuery = "mobile=13800138000"
        controller.GetByMobile(ctx)
        assert.Equal(t, http.StatusOK, rr.Code)
        resp := decodeJSON(rr)
        assert.Equal(t, "13800138000", resp["mobile"])
        assert.Equal(t, "Tom", resp["name"])
    })

    t.Run("lookup requires mobile", func(t *testing.T) {
        ctx, rr := newJSONContext(http.MethodGet, "/v1/user/get", nil)
        controller.GetByMobile(ctx)
        assert.Equal(t, http.StatusBadRequest, rr.Code)
        resp := decodeJSON(rr)
        assert.Equal(t, "mobile parameter is required", resp["msg"])
    })

    t.Run("get by id invalid", func(t *testing.T) {
        ctx, rr := newJSONContext(http.MethodGet, "/v1/user/get", nil)
        ctx.Request.URL.RawQuery = "id=abc"
        controller.GetById(ctx)
        assert.Equal(t, http.StatusBadRequest, rr.Code)
        resp := decodeJSON(rr)
        assert.Equal(t, "invalid id parameter", resp["msg"])
    })

    t.Run("get by id success", func(t *testing.T) {
        expected := &uv1.UserDTO{User: uv1.User{ID: 45, Mobile: "13800138000", NickName: "Jerry", Gender: "male", Birthday: itime.Time{Time: time.Date(1993, 4, 5, 0, 0, 0, 0, time.UTC)}}}
        userSvc.getFunc = func(ctx context.Context, id uint64) (*uv1.UserDTO, error) { assert.Equal(t, uint64(45), id); return expected, nil }
        ctx, rr := newJSONContext(http.MethodGet, "/v1/user/get", nil)
        ctx.Request.URL.RawQuery = "id=45"
        controller.GetById(ctx)
        assert.Equal(t, http.StatusOK, rr.Code)
        resp := decodeJSON(rr)
        assert.Equal(t, float64(expected.ID), resp["id"])
        assert.Equal(t, expected.NickName, resp["name"])
    })
}

func TestUserController_List(t *testing.T) {
    userSvc := &fakeUserService{}
    controller := NewUserController(&fakeTranslator{}, &fakeServiceFactory{user: userSvc})
    t.Run("paginated list", func(t *testing.T) {
        userSvc.getUserListFunc = func(ctx context.Context, pn, pSize uint32) (*uv1.UserListDTO, error) {
            assert.Equal(t, uint32(2), pn)
            assert.Equal(t, uint32(5), pSize)
            return &uv1.UserListDTO{TotalCount: 12, Items: []*uv1.UserDTO{
                {User: uv1.User{ID: 1, NickName: "A", Mobile: "1", Gender: "male", Birthday: itime.Time{Time: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)}}},
                {User: uv1.User{ID: 2, NickName: "B", Mobile: "2", Gender: "female", Birthday: itime.Time{Time: time.Date(1991, 2, 2, 0, 0, 0, 0, time.UTC)}}},
            }}, nil
        }
        ctx, rr := newJSONContext(http.MethodGet, "/v1/user/list", nil)
        ctx.Request.URL.RawQuery = "pn=2&pSize=5"
        controller.GetUserList(ctx)
        assert.Equal(t, http.StatusOK, rr.Code)
        resp := decodeJSON(rr)
        assert.Equal(t, float64(12), resp["total"])
        users, ok := resp["users"].([]any)
        assert.True(t, ok)
        assert.Len(t, users, 2)
    })

    t.Run("invalid pn", func(t *testing.T) {
        ctx, rr := newJSONContext(http.MethodGet, "/v1/user/list", nil)
        ctx.Request.URL.RawQuery = "pn=abc&pSize=5"
        controller.GetUserList(ctx)
        assert.Equal(t, http.StatusBadRequest, rr.Code)
        resp := decodeJSON(rr)
        assert.Equal(t, "invalid pn parameter", resp["msg"])
    })
}

// Helpers and fakes ---------------------------------------------------------

type fakeTranslator struct{ messages map[string]string }
func (t *fakeTranslator) T(key string, params ...interface{}) string {
    if t == nil { return key }
    if msg, ok := t.messages[key]; ok {
        switch len(params) {
        case 0:
            return msg
        case 1:
            return fmt.Sprintf(msg, params[0])
        default:
            return fmt.Sprintf(msg, params[0], params[1])
        }
    }
    return key
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
func (f *fakeUserService) MobileLogin(ctx context.Context, mobile, password string) (*uv1.UserDTO, error) { return f.mobileLoginFunc(ctx, mobile, password) }
func (f *fakeUserService) Register(ctx context.Context, mobile, password, code string) (*uv1.UserDTO, error) { return f.registerFunc(ctx, mobile, password, code) }
func (f *fakeUserService) Update(ctx context.Context, userDTO *uv1.UserDTO) error { return f.updateFunc(ctx, userDTO) }
func (f *fakeUserService) Get(ctx context.Context, userID uint64) (*uv1.UserDTO, error) { return f.getFunc(ctx, userID) }
func (f *fakeUserService) GetByMobile(ctx context.Context, mobile string) (*uv1.UserDTO, error) { return f.getByMobileFunc(ctx, mobile) }
func (f *fakeUserService) GetUserList(ctx context.Context, pn, pSize uint32) (*uv1.UserListDTO, error) { return f.getUserListFunc(ctx, pn, pSize) }
func (f *fakeUserService) CheckPassWord(ctx context.Context, password, encrypted string) (bool, error) { return f.checkPasswordFunc(ctx, password, encrypted) }

type fakeServiceFactory struct{ user uv1.UserSrv }
func (f *fakeServiceFactory) Goods() gv1.GoodsSrv         { return nil }
func (f *fakeServiceFactory) Users() uv1.UserSrv          { return f.user }
func (f *fakeServiceFactory) Sms() sv1.SmsSrv             { return nil }
func (f *fakeServiceFactory) Inventory() iv1.InventorySrv { return nil }
func (f *fakeServiceFactory) Order() ov1.OrderSrv         { return nil }
func (f *fakeServiceFactory) UserOp() uopv1.UserOpSrv     { return nil }
func (f *fakeServiceFactory) Coupon() cv1.CouponSrv       { return nil }
func (f *fakeServiceFactory) Payment() pv1.PaymentSrv     { return nil }
func (f *fakeServiceFactory) Logistics() lv1.LogisticsSrv { return nil }

type captchaStore interface { Set(id string, value string) error; Get(id string, clear bool) string; Verify(id, answer string, clear bool) bool }
type fakeCaptchaStore struct{ expectedID, expectedAnswer string; verifyResult bool }
func (f *fakeCaptchaStore) Set(id string, value string) error { return nil }
func (f *fakeCaptchaStore) Get(id string, clear bool) string  { return "" }
func (f *fakeCaptchaStore) Verify(id, answer string, clear bool) bool { return f.verifyResult }

func newJSONContext(method, path string, payload interface{}) (*gin.Context, *httptest.ResponseRecorder) {
    var body []byte
    var err error
    if payload != nil { body, err = json.Marshal(payload); if err != nil { panic(err) } }
    req, err := http.NewRequest(method, path, bytes.NewReader(body))
    if err != nil { panic(err) }
    if payload != nil { req.Header.Set("Content-Type", "application/json") }
    rr := httptest.NewRecorder()
    ctx, _ := gin.CreateTestContext(rr)
    ctx.Request = req
    return ctx, rr
}

func decodeJSON(rr *httptest.ResponseRecorder) map[string]interface{} {
    var resp map[string]interface{}
    if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil { panic(fmt.Errorf("failed to decode response: %w", err)) }
    return resp
}
