package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080/api/v1"

func waitForAPI(t *testing.T) {
	for i := 0; i < 30; i++ {
		resp, err := http.Get(baseURL + "/choose_banner")
		if err == nil {
			require.NoError(t, resp.Body.Close())
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("API not available after 30s")
}

func TestE2E_BannerRotation(t *testing.T) {
	if os.Getenv("E2E") == "" {
		t.Skip("E2E env not set")
	}
	waitForAPI(t)

	// 1. Добавить баннер в слот
	addReq := map[string]interface{}{"slot_id": 1, "banner_id": 100}
	body, err := json.Marshal(addReq)
	require.NoError(t, err)
	resp, err := http.Post(baseURL+"/banner_slot", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	// 2. Выбрать баннер для показа
	chooseReq := map[string]interface{}{"slot_id": 1, "group_id": 1}
	body, err = json.Marshal(chooseReq)
	require.NoError(t, err)
	resp, err = http.Post(baseURL+"/choose_banner", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var chooseResp map[string]interface{}
	require.NoError(t, json.Unmarshal(respBody, &chooseResp))
	require.Equal(t, float64(100), chooseResp["banner_id"])

	// 3. Засчитать клик
	clickReq := map[string]interface{}{"slot_id": 1, "banner_id": 100, "group_id": 1}
	body, err = json.Marshal(clickReq)
	require.NoError(t, err)
	resp, err = http.Post(baseURL+"/register_click", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	// 4. Удалить баннер - ИСПРАВЛЕНО
	deleteReq := map[string]interface{}{"slot_id": 1, "banner_id": 100}
	body, err = json.Marshal(deleteReq)
	require.NoError(t, err)
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", baseURL+"/banner_slot", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	// 5. Попробовать выбрать баннер (должна быть ошибка)
	body, err = json.Marshal(chooseReq)
	require.NoError(t, err)
	resp, err = http.Post(baseURL+"/choose_banner", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.GreaterOrEqual(t, resp.StatusCode, 400)
}
