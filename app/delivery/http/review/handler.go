package http_review

import (
	"net/http"

	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	request_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/request"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/gin-gonic/gin"
)

type routeHandler struct {
	Usecase    domain.ReviewAppUsecase
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewReviewHandler(r *gin.RouterGroup, mdl middleware.Middleware, uc domain.ReviewAppUsecase) {
	handler := &routeHandler{
		Usecase:    uc,
		Route:      r,
		Middleware: mdl,
	}

	api := r.Group("/review", mdl.Auth())
	api.GET("/question", mdl.AuthRole("Employee", "Admin"), handler.FetchQuestions)
	api.POST("", mdl.AuthEmployee(), handler.CreateOrUpdate)
	api.GET("/:id", mdl.AuthRole("Employee", "Admin"), handler.GetByID)
}

// Fetch Review Questions
// @Summary Fetch review questions
// @Description Get the list of review questions for the employee review form
// @Tags Review
// @Produce json
// @Success 200 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /review/question [get]
// @Security BearerAuth
//
// Example success response:
// {
//   "status": 200,
//   "message": "success",
//   "validation": null,
//   "data": [
//     {
//       "id": "11111111-1111-1111-1111-111111111101",
//       "indicator": "WORK_QUALITY",
//       "question_text": "Sejauh mana hasil pekerjaan freelancer sesuai dengan requirement dan spesifikasi yang telah ditentukan?",
//       "question_order": 1
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111102",
//       "indicator": "WORK_QUALITY",
//       "question_text": "Seberapa rendah tingkat bug atau kesalahan yang ditemukan pada hasil pekerjaan freelancer?",
//       "question_order": 2
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111103",
//       "indicator": "WORK_QUALITY",
//       "question_text": "Seberapa rapi dan mudah dipahami struktur hasil kerja yang dihasilkan freelancer?",
//       "question_order": 3
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111104",
//       "indicator": "WORK_QUALITY",
//       "question_text": "Sejauh mana hasil akhir pekerjaan freelancer memenuhi standar kualitas yang ditetapkan perusahaan?",
//       "question_order": 4
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111105",
//       "indicator": "TIMELINESS",
//       "question_text": "Sejauh mana freelancer menyelesaikan tugas sesuai dengan deadline yang telah disepakati?",
//       "question_order": 1
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111106",
//       "indicator": "TIMELINESS",
//       "question_text": "Seberapa cepat freelancer menindaklanjuti revisi atau permintaan perbaikan yang diberikan?",
//       "question_order": 2
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111107",
//       "indicator": "TIMELINESS",
//       "question_text": "Sejauh mana freelancer mampu mengatur prioritas kerja ketika menangani beberapa tugas sekaligus?",
//       "question_order": 3
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111108",
//       "indicator": "TIMELINESS",
//       "question_text": "Seberapa jarang freelancer mengalami keterlambatan tanpa alasan yang jelas?",
//       "question_order": 4
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111109",
//       "indicator": "COMMUNICATION_COLLABORATION",
//       "question_text": "Sejauh mana freelancer responsif dalam berkomunikasi melalui chat, email, atau meeting?",
//       "question_order": 1
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111110",
//       "indicator": "COMMUNICATION_COLLABORATION",
//       "question_text": "Seberapa proaktif freelancer mengajukan pertanyaan ketika terdapat instruksi atau kebutuhan yang belum jelas?",
//       "question_order": 2
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111111",
//       "indicator": "COMMUNICATION_COLLABORATION",
//       "question_text": "Sejauh mana freelancer terbuka dan menerima feedback atau saran perbaikan yang diberikan?",
//       "question_order": 3
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111112",
//       "indicator": "COMMUNICATION_COLLABORATION",
//       "question_text": "Seberapa baik freelancer dapat bekerja sama dan berkoordinasi dengan anggota tim lainnya?",
//       "question_order": 4
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111113",
//       "indicator": "PROBLEM_SOLVING_INITIATIVE",
//       "question_text": "Sejauh mana freelancer berusaha mencari solusi secara mandiri sebelum meminta bantuan?",
//       "question_order": 1
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111114",
//       "indicator": "PROBLEM_SOLVING_INITIATIVE",
//       "question_text": "Seberapa besar inisiatif freelancer dalam mengambil tindakan untuk menyelesaikan masalah yang muncul?",
//       "question_order": 2
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111115",
//       "indicator": "PROBLEM_SOLVING_INITIATIVE",
//       "question_text": "Sejauh mana freelancer memberikan saran atau improvement yang relevan terhadap project yang dikerjakan?",
//       "question_order": 3
//     },
//     {
//       "id": "11111111-1111-1111-1111-111111111116",
//       "indicator": "PROBLEM_SOLVING_INITIATIVE",
//       "question_text": "Seberapa baik freelancer menangani kendala teknis secara tenang, sistematis, dan terstruktur?",
//       "question_order": 4
//     }
//   ]
// }
func (h *routeHandler) FetchQuestions(c *gin.Context) {
	res := h.Usecase.FetchQuestions(c.Request.Context())
	c.JSON(res.Status, res)
}

// Create or Update Review
// @Summary Create or update a review for a finished onboarding
// @Description Employee creates or updates a review for a candidate that has finished onboarding
// @Tags Review
// @Accept json
// @Produce json
// @Param request body request_model.CreateReviewRequest true "Review Request"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /review [post]
// @Security BearerAuth
//
// Example success response:
// {
//   "status": 200,
//   "message": "success",
//   "validation": null,
//   "data": {
//     "id": "9b8d7f53-7c7e-4df7-b4c1-3e6f8f0a1c21",
//     "subrequest_id": "8b9d6d22-7c6a-4cb2-9f3f-9d2bc3a7d111",
//     "candidate_user_id": "2f3d4a55-0d67-4df0-9a3f-7c9f6a2d1234",
//     "employee_user_id": "b7f3d9a1-4f2a-4d3b-8b1d-1f77f7a8c222",
//     "onboard_history_id": "4c1c9d77-8b33-4c2b-9a1c-6f7d2b3a4444",
//     "final_recommendation": "RECOMMENDED",
//     "notes": "Candidate communicated well, delivered on time, and handled issues independently.",
//     "answers": [
//       {
//         "question_id": "11111111-1111-1111-1111-111111111101",
//         "indicator": "WORK_QUALITY",
//         "question_text": "Sejauh mana hasil pekerjaan freelancer sesuai dengan requirement dan spesifikasi yang telah ditentukan?",
//         "question_order": 1,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111102",
//         "indicator": "WORK_QUALITY",
//         "question_text": "Seberapa rendah tingkat bug atau kesalahan yang ditemukan pada hasil pekerjaan freelancer?",
//         "question_order": 2,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111103",
//         "indicator": "WORK_QUALITY",
//         "question_text": "Seberapa rapi dan mudah dipahami struktur hasil kerja yang dihasilkan freelancer?",
//         "question_order": 3,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111104",
//         "indicator": "WORK_QUALITY",
//         "question_text": "Sejauh mana hasil akhir pekerjaan freelancer memenuhi standar kualitas yang ditetapkan perusahaan?",
//         "question_order": 4,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111105",
//         "indicator": "TIMELINESS",
//         "question_text": "Sejauh mana freelancer menyelesaikan tugas sesuai dengan deadline yang telah disepakati?",
//         "question_order": 1,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111106",
//         "indicator": "TIMELINESS",
//         "question_text": "Seberapa cepat freelancer menindaklanjuti revisi atau permintaan perbaikan yang diberikan?",
//         "question_order": 2,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111107",
//         "indicator": "TIMELINESS",
//         "question_text": "Sejauh mana freelancer mampu mengatur prioritas kerja ketika menangani beberapa tugas sekaligus?",
//         "question_order": 3,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111108",
//         "indicator": "TIMELINESS",
//         "question_text": "Seberapa jarang freelancer mengalami keterlambatan tanpa alasan yang jelas?",
//         "question_order": 4,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111109",
//         "indicator": "COMMUNICATION_COLLABORATION",
//         "question_text": "Sejauh mana freelancer responsif dalam berkomunikasi melalui chat, email, atau meeting?",
//         "question_order": 1,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111110",
//         "indicator": "COMMUNICATION_COLLABORATION",
//         "question_text": "Seberapa proaktif freelancer mengajukan pertanyaan ketika terdapat instruksi atau kebutuhan yang belum jelas?",
//         "question_order": 2,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111111",
//         "indicator": "COMMUNICATION_COLLABORATION",
//         "question_text": "Sejauh mana freelancer terbuka dan menerima feedback atau saran perbaikan yang diberikan?",
//         "question_order": 3,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111112",
//         "indicator": "COMMUNICATION_COLLABORATION",
//         "question_text": "Seberapa baik freelancer dapat bekerja sama dan berkoordinasi dengan anggota tim lainnya?",
//         "question_order": 4,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111113",
//         "indicator": "PROBLEM_SOLVING_INITIATIVE",
//         "question_text": "Sejauh mana freelancer berusaha mencari solusi secara mandiri sebelum meminta bantuan?",
//         "question_order": 1,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111114",
//         "indicator": "PROBLEM_SOLVING_INITIATIVE",
//         "question_text": "Seberapa besar inisiatif freelancer dalam mengambil tindakan untuk menyelesaikan masalah yang muncul?",
//         "question_order": 2,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111115",
//         "indicator": "PROBLEM_SOLVING_INITIATIVE",
//         "question_text": "Sejauh mana freelancer memberikan saran atau improvement yang relevan terhadap project yang dikerjakan?",
//         "question_order": 3,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111116",
//         "indicator": "PROBLEM_SOLVING_INITIATIVE",
//         "question_text": "Seberapa baik freelancer menangani kendala teknis secara tenang, sistematis, dan terstruktur?",
//         "question_order": 4,
//         "score": 5
//       }
//     ],
//     "created_at": "2026-05-29T21:45:12Z",
//     "updated_at": "2026-05-29T21:45:12Z"
//   }
// }
func (h *routeHandler) CreateOrUpdate(c *gin.Context) {
	claims, ok := c.Get("token_data")
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized"))
		return
	}

	var req request_model.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid request body"))
		return
	}

	employeeID := claims.(domain.JWTClaimUser).UserID
	res := h.Usecase.CreateOrUpdate(c.Request.Context(), employeeID, req)
	c.JSON(res.Status, res)
}

// Get Review By ID
// @Summary Get review details
// @Description Fetch review details by review ID (Employee/Admin)
// @Tags Review
// @Produce json
// @Param id path string true "Review ID"
// @Success 200 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 401 {object} response.Base
// @Failure 403 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /review/{id} [get]
// @Security BearerAuth
//
// Example success response:
// {
//   "status": 200,
//   "message": "success",
//   "validation": null,
//   "data": {
//     "id": "9b8d7f53-7c7e-4df7-b4c1-3e6f8f0a1c21",
//     "subrequest_id": "8b9d6d22-7c6a-4cb2-9f3f-9d2bc3a7d111",
//     "candidate_user_id": "2f3d4a55-0d67-4df0-9a3f-7c9f6a2d1234",
//     "employee_user_id": "b7f3d9a1-4f2a-4d3b-8b1d-1f77f7a8c222",
//     "onboard_history_id": "4c1c9d77-8b33-4c2b-9a1c-6f7d2b3a4444",
//     "final_recommendation": "RECOMMENDED",
//     "notes": "Candidate communicated well, delivered on time, and handled issues independently.",
//     "answers": [
//       {
//         "question_id": "11111111-1111-1111-1111-111111111101",
//         "indicator": "WORK_QUALITY",
//         "question_text": "Sejauh mana hasil pekerjaan freelancer sesuai dengan requirement dan spesifikasi yang telah ditentukan?",
//         "question_order": 1,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111102",
//         "indicator": "WORK_QUALITY",
//         "question_text": "Seberapa rendah tingkat bug atau kesalahan yang ditemukan pada hasil pekerjaan freelancer?",
//         "question_order": 2,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111103",
//         "indicator": "WORK_QUALITY",
//         "question_text": "Seberapa rapi dan mudah dipahami struktur hasil kerja yang dihasilkan freelancer?",
//         "question_order": 3,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111104",
//         "indicator": "WORK_QUALITY",
//         "question_text": "Sejauh mana hasil akhir pekerjaan freelancer memenuhi standar kualitas yang ditetapkan perusahaan?",
//         "question_order": 4,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111105",
//         "indicator": "TIMELINESS",
//         "question_text": "Sejauh mana freelancer menyelesaikan tugas sesuai dengan deadline yang telah disepakati?",
//         "question_order": 1,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111106",
//         "indicator": "TIMELINESS",
//         "question_text": "Seberapa cepat freelancer menindaklanjuti revisi atau permintaan perbaikan yang diberikan?",
//         "question_order": 2,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111107",
//         "indicator": "TIMELINESS",
//         "question_text": "Sejauh mana freelancer mampu mengatur prioritas kerja ketika menangani beberapa tugas sekaligus?",
//         "question_order": 3,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111108",
//         "indicator": "TIMELINESS",
//         "question_text": "Seberapa jarang freelancer mengalami keterlambatan tanpa alasan yang jelas?",
//         "question_order": 4,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111109",
//         "indicator": "COMMUNICATION_COLLABORATION",
//         "question_text": "Sejauh mana freelancer responsif dalam berkomunikasi melalui chat, email, atau meeting?",
//         "question_order": 1,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111110",
//         "indicator": "COMMUNICATION_COLLABORATION",
//         "question_text": "Seberapa proaktif freelancer mengajukan pertanyaan ketika terdapat instruksi atau kebutuhan yang belum jelas?",
//         "question_order": 2,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111111",
//         "indicator": "COMMUNICATION_COLLABORATION",
//         "question_text": "Sejauh mana freelancer terbuka dan menerima feedback atau saran perbaikan yang diberikan?",
//         "question_order": 3,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111112",
//         "indicator": "COMMUNICATION_COLLABORATION",
//         "question_text": "Seberapa baik freelancer dapat bekerja sama dan berkoordinasi dengan anggota tim lainnya?",
//         "question_order": 4,
//         "score": 5
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111113",
//         "indicator": "PROBLEM_SOLVING_INITIATIVE",
//         "question_text": "Sejauh mana freelancer berusaha mencari solusi secara mandiri sebelum meminta bantuan?",
//         "question_order": 1,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111114",
//         "indicator": "PROBLEM_SOLVING_INITIATIVE",
//         "question_text": "Seberapa besar inisiatif freelancer dalam mengambil tindakan untuk menyelesaikan masalah yang muncul?",
//         "question_order": 2,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111115",
//         "indicator": "PROBLEM_SOLVING_INITIATIVE",
//         "question_text": "Sejauh mana freelancer memberikan saran atau improvement yang relevan terhadap project yang dikerjakan?",
//         "question_order": 3,
//         "score": 4
//       },
//       {
//         "question_id": "11111111-1111-1111-1111-111111111116",
//         "indicator": "PROBLEM_SOLVING_INITIATIVE",
//         "question_text": "Seberapa baik freelancer menangani kendala teknis secara tenang, sistematis, dan terstruktur?",
//         "question_order": 4,
//         "score": 5
//       }
//     ],
//     "created_at": "2026-05-29T21:45:12Z",
//     "updated_at": "2026-05-29T21:45:12Z"
//   }
// }
func (h *routeHandler) GetByID(c *gin.Context) {
	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "Invalid id parameter"))
		return
	}

	res := h.Usecase.FetchByID(c.Request.Context(), reviewID)
	c.JSON(res.Status, res)
}
