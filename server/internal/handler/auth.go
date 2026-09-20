package handler

import (
	"context"
	"errors"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/generated/oapi"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/repository"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/requestctx"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/service"
)

type AuthUseCases interface {
	Login(context.Context, string, string, string, string) (service.AuthResult, error)
	Refresh(context.Context, string, string, string) (service.AuthResult, error)
	Logout(context.Context, string) error
	CurrentUser(context.Context, int64) (repository.User, error)
}

type Handler struct {
	auth    AuthUseCases
	cookies security.CookieManager
	now     func() time.Time
}

func NewAuth(auth AuthUseCases, cookies security.CookieManager) *Handler {
	return &Handler{auth: auth, cookies: cookies, now: time.Now}
}

func (h *Handler) Login(ctx context.Context, request oapi.LoginRequestObject) (oapi.LoginResponseObject, error) {
	if request.Body == nil {
		return oapi.Login400JSONResponse{BadRequestJSONResponse: badRequest(40000, "request body is required")}, nil
	}
	metadata := requestctx.MetadataFrom(ctx)
	result, err := h.auth.Login(ctx, request.Body.Username, request.Body.Password, metadata.UserAgent, metadata.IPHash)
	if err != nil {
		if errors.Is(err, domain.ErrBadRequest) {
			return oapi.Login400JSONResponse{BadRequestJSONResponse: badRequest(40000, "invalid login request")}, nil
		}
		if errors.Is(err, domain.ErrUnauthorized) || errors.Is(err, domain.ErrNotFound) {
			return oapi.Login401JSONResponse{UnauthorizedJSONResponse: unauthorized(40100, "invalid credentials")}, nil
		}
		return nil, err
	}
	response := oapi.Login200JSONResponse{}
	response.Body.Code = 0
	response.Body.Message = "success"
	response.Body.Data = &oapi.AccessToken{AccessToken: result.AccessToken, TokenType: "Bearer", ExpiresIn: result.ExpiresIn}
	response.Headers.SetCookie = h.cookies.Set(result.RefreshToken, h.now().UTC())
	return response, nil
}

func (h *Handler) RefreshAccessToken(ctx context.Context, _ oapi.RefreshAccessTokenRequestObject) (oapi.RefreshAccessTokenResponseObject, error) {
	metadata := requestctx.MetadataFrom(ctx)
	result, err := h.auth.Refresh(ctx, requestctx.RefreshTokenFrom(ctx), metadata.UserAgent, metadata.IPHash)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) || errors.Is(err, domain.ErrReplay) {
			return oapi.RefreshAccessToken401JSONResponse{UnauthorizedJSONResponse: unauthorized(40100, "session is invalid")}, nil
		}
		return nil, err
	}
	response := oapi.RefreshAccessToken200JSONResponse{}
	response.Body.Code = 0
	response.Body.Message = "success"
	response.Body.Data = &oapi.AccessToken{AccessToken: result.AccessToken, TokenType: "Bearer", ExpiresIn: result.ExpiresIn}
	response.Headers.SetCookie = h.cookies.Set(result.RefreshToken, h.now().UTC())
	return response, nil
}

func (h *Handler) Logout(ctx context.Context, _ oapi.LogoutRequestObject) (oapi.LogoutResponseObject, error) {
	if err := h.auth.Logout(ctx, requestctx.RefreshTokenFrom(ctx)); err != nil && !errors.Is(err, domain.ErrUnauthorized) {
		return nil, err
	}
	return oapi.Logout200JSONResponse{
		Body:    oapi.APIResponse{Code: 0, Message: "success"},
		Headers: oapi.Logout200ResponseHeaders{SetCookie: h.cookies.Clear()},
	}, nil
}

func (h *Handler) GetCurrentUser(ctx context.Context, _ oapi.GetCurrentUserRequestObject) (oapi.GetCurrentUserResponseObject, error) {
	identity, ok := requestctx.IdentityFrom(ctx)
	if !ok {
		return oapi.GetCurrentUser401JSONResponse{UnauthorizedJSONResponse: unauthorized(40100, "authentication is required")}, nil
	}
	user, err := h.auth.CurrentUser(ctx, identity.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrUnauthorized) {
			return oapi.GetCurrentUser401JSONResponse{UnauthorizedJSONResponse: unauthorized(40100, "authentication is required")}, nil
		}
		return nil, err
	}
	return oapi.GetCurrentUser200JSONResponse{Code: 0, Message: "success", Data: userDTO(user)}, nil
}

func errorResponse(code int, message string) oapi.ErrorResponse {
	return oapi.ErrorResponse{Code: code, Message: message, Data: nil}
}

// badRequest / unauthorized 把统一的 ErrorResponse 转成各状态码专用的具名响应体类型。
// 生成的类型是 `type BadRequestJSONResponse ErrorResponse` 这样的具名类型，
// 与 ErrorResponse 之间不能隐式赋值，必须显式转换。
func badRequest(code int, message string) oapi.BadRequestJSONResponse {
	return oapi.BadRequestJSONResponse(errorResponse(code, message))
}

func unauthorized(code int, message string) oapi.UnauthorizedJSONResponse {
	return oapi.UnauthorizedJSONResponse(errorResponse(code, message))
}

func userDTO(user repository.User) *oapi.User {
	return &oapi.User{
		ID: user.ID, Username: user.Username, Nickname: user.Nickname,
		Email: user.Email, AvatarURL: user.AvatarURL, Bio: user.Bio,
		Role: oapi.UserRole(user.Role), CreatedAt: user.CreatedAt,
	}
}
