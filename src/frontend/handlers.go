// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/genproto"
	"github.com/GoogleCloudPlatform/microservices-demo/src/frontend/money"
)

var (
	templates = template.Must(template.New("").
		Funcs(template.FuncMap{
			"renderMoney": renderMoney,
		}).ParseGlob("templates/*.html"))
)

func (fe *frontendServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.WithField("currency", currentCurrency(r)).Info("home")
	currencies, err := fe.getCurrencies(r.Context())
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}
	products, err := fe.getProducts(r.Context())
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve products"), http.StatusInternalServerError)
		return
	}
	cart, err := fe.getCart(r.Context(), sessionID(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve cart"), http.StatusInternalServerError)
		return
	}

	type productView struct {
		Item  *pb.Product
		Price *pb.Money
	}
	ps := make([]productView, len(products))
	for i, p := range products {
		price, err := fe.convertCurrency(r.Context(), p.GetPriceUsd(), currentCurrency(r))
		if err != nil {
			renderHTTPError(log, r, w, errors.Wrapf(err, "failed to do currency conversion for product %s", p.GetId()), http.StatusInternalServerError)
			return
		}
		ps[i] = productView{p, price}
	}

	if err := templates.ExecuteTemplate(w, "home", map[string]interface{}{
		"session_id":    sessionID(r),
		"request_id":    r.Context().Value(ctxKeyRequestID{}),
		"user_currency": currentCurrency(r),
		"show_currency": true,
		"currencies":    currencies,
		"products":      ps,
		"cart_size":     cartSize(cart), // illustrates canary deployments
		"ad":            fe.chooseAd(r.Context(), []string{}, log),
		"loggedin":      loggedin,
		"username":      user,
	}); err != nil {
		log.Error(err)
	}
}

func (fe *frontendServer) productHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	id := mux.Vars(r)["id"]
	if id == "" {
		renderHTTPError(log, r, w, errors.New("product id not specified"), http.StatusBadRequest)
		return
	}
	log.WithField("id", id).WithField("currency", currentCurrency(r)).
		Debug("serving product page")

	p, err := fe.getProduct(r.Context(), id)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve product"), http.StatusInternalServerError)
		return
	}
	currencies, err := fe.getCurrencies(r.Context())
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}

	cart, err := fe.getCart(r.Context(), sessionID(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve cart"), http.StatusInternalServerError)
		return
	}

	price, err := fe.convertCurrency(r.Context(), p.GetPriceUsd(), currentCurrency(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to convert currency"), http.StatusInternalServerError)
		return
	}

	recommendations, err := fe.getRecommendations(r.Context(), sessionID(r), []string{id})
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to get product recommendations"), http.StatusInternalServerError)
		return
	}

	product := struct {
		Item  *pb.Product
		Price *pb.Money
	}{p, price}

	if err := templates.ExecuteTemplate(w, "product", map[string]interface{}{
		"session_id":      sessionID(r),
		"request_id":      r.Context().Value(ctxKeyRequestID{}),
		"ad":              fe.chooseAd(r.Context(), p.Categories, log),
		"user_currency":   currentCurrency(r),
		"show_currency":   true,
		"currencies":      currencies,
		"product":         product,
		"recommendations": recommendations,
		"cart_size":       cartSize(cart),
		"loggedin":        loggedin,
		"username":        user,
	}); err != nil {
		log.Println(err)
	}
}

func (fe *frontendServer) addToCartHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	quantity, _ := strconv.ParseUint(r.FormValue("quantity"), 10, 32)
	productID := r.FormValue("product_id")
	if productID == "" || quantity == 0 {
		renderHTTPError(log, r, w, errors.New("invalid form input"), http.StatusBadRequest)
		return
	}
	log.WithField("product", productID).WithField("quantity", quantity).Debug("adding to cart")

	p, err := fe.getProduct(r.Context(), productID)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve product"), http.StatusInternalServerError)
		return
	}

	if err := fe.insertCart(r.Context(), sessionID(r), p.GetId(), int32(quantity)); err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to add to cart"), http.StatusInternalServerError)
		return
	}
	w.Header().Set("location", "/cart")
	w.WriteHeader(http.StatusFound)
}

func (fe *frontendServer) emptyCartHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.Debug("emptying cart")

	if err := fe.emptyCart(r.Context(), sessionID(r)); err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to empty cart"), http.StatusInternalServerError)
		return
	}
	w.Header().Set("location", "/")
	w.WriteHeader(http.StatusFound)
}

func (fe *frontendServer) viewCartHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.Debug("view user cart")
	currencies, err := fe.getCurrencies(r.Context())
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}
	cart, err := fe.getCart(r.Context(), sessionID(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve cart"), http.StatusInternalServerError)
		return
	}

	recommendations, err := fe.getRecommendations(r.Context(), sessionID(r), cartIDs(cart))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to get product recommendations"), http.StatusInternalServerError)
		return
	}

	shippingCost, err := fe.getShippingQuote(r.Context(), cart, currentCurrency(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to get shipping quote"), http.StatusInternalServerError)
		return
	}

	type cartItemView struct {
		Item     *pb.Product
		Quantity int32
		Price    *pb.Money
	}
	items := make([]cartItemView, len(cart))
	totalPrice := pb.Money{CurrencyCode: currentCurrency(r)}
	for i, item := range cart {
		p, err := fe.getProduct(r.Context(), item.GetProductId())
		if err != nil {
			renderHTTPError(log, r, w, errors.Wrapf(err, "could not retrieve product #%s", item.GetProductId()), http.StatusInternalServerError)
			return
		}
		price, err := fe.convertCurrency(r.Context(), p.GetPriceUsd(), currentCurrency(r))
		if err != nil {
			renderHTTPError(log, r, w, errors.Wrapf(err, "could not convert currency for product #%s", item.GetProductId()), http.StatusInternalServerError)
			return
		}

		multPrice := money.MultiplySlow(*price, uint32(item.GetQuantity()))
		items[i] = cartItemView{
			Item:     p,
			Quantity: item.GetQuantity(),
			Price:    &multPrice}
		totalPrice = money.Must(money.Sum(totalPrice, multPrice))
	}
	totalPrice = money.Must(money.Sum(totalPrice, *shippingCost))

	year := time.Now().Year()
	if err := templates.ExecuteTemplate(w, "cart", map[string]interface{}{
		"session_id":       sessionID(r),
		"request_id":       r.Context().Value(ctxKeyRequestID{}),
		"user_currency":    currentCurrency(r),
		"currencies":       currencies,
		"recommendations":  recommendations,
		"cart_size":        cartSize(cart),
		"shipping_cost":    shippingCost,
		"show_currency":    true,
		"total_cost":       totalPrice,
		"items":            items,
		"expiration_years": []int{year, year + 1, year + 2, year + 3, year + 4, year + 5, year + 6, year + 7, year + 8},
		"loggedin":         loggedin,
		"username":         user,
	}); err != nil {
		log.Println(err)
	}
}

func (fe *frontendServer) placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.Debug("placing order")
	gofakeit.Seed(0)

	var (
		email         = gofakeit.Email()
		streetAddress = fmt.Sprintf("%s %s", gofakeit.StreetNumber(), gofakeit.StreetName())
		zipCode, _    = strconv.Atoi(gofakeit.Zip())
		city          = gofakeit.City()
		state         = gofakeit.State()
		country       = gofakeit.Country()
		ccNumber      = r.FormValue("credit_card_number")
		ccMonth, _    = strconv.ParseInt(r.FormValue("credit_card_expiration_month"), 10, 32)
		ccYear, _     = strconv.ParseInt(r.FormValue("credit_card_expiration_year"), 10, 32)
		ccCVV, _      = strconv.Atoi(gofakeit.CreditCardCvv())
		save_info     = r.FormValue("save_credit_card")
	)

	cc_split := strings.Split(ccNumber, "-")
	for i := 0; i < 3; i++ {
		ccNumber = strings.Replace(ccNumber, cc_split[i], randNum(), 1)
	}
	fmt.Println(ccNumber)

	if save_info != "" {
		data := map[string]string{
			"credit_card_number": ccNumber,
			"username":           user,
		}

		json_data, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := http.Post("http://"+fe.userDbSvcAddr+"/add_credit_card", "application/json", bytes.NewBuffer(json_data))
		if err != nil {
			log.Fatal(err)
			renderHTTPError(log, r, w, errors.Wrap(err, "Error logging in!"), http.StatusBadRequest)
		}
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Println(res)
	}

	order, err := pb.NewCheckoutServiceClient(fe.checkoutSvcConn).
		PlaceOrder(r.Context(), &pb.PlaceOrderRequest{
			Email: email,
			CreditCard: &pb.CreditCardInfo{
				CreditCardNumber:          ccNumber,
				CreditCardExpirationMonth: int32(ccMonth),
				CreditCardExpirationYear:  int32(ccYear),
				CreditCardCvv:             int32(ccCVV)},
			UserId:       sessionID(r),
			UserCurrency: currentCurrency(r),
			Address: &pb.Address{
				StreetAddress: streetAddress,
				City:          city,
				State:         state,
				ZipCode:       int32(zipCode),
				Country:       country},
		})
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to complete the order"), http.StatusInternalServerError)
		return
	}
	log.WithField("order", order.GetOrder().GetOrderId()).Info("order placed")

	order.GetOrder().GetItems()
	recommendations, _ := fe.getRecommendations(r.Context(), sessionID(r), nil)

	totalPaid := *order.GetOrder().GetShippingCost()
	for _, v := range order.GetOrder().GetItems() {
		multPrice := money.MultiplySlow(*v.GetCost(), uint32(v.GetItem().GetQuantity()))
		totalPaid = money.Must(money.Sum(totalPaid, multPrice))
	}

	currencies, err := fe.getCurrencies(r.Context())
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "order", map[string]interface{}{
		"session_id":      sessionID(r),
		"request_id":      r.Context().Value(ctxKeyRequestID{}),
		"user_currency":   currentCurrency(r),
		"show_currency":   false,
		"currencies":      currencies,
		"order":           order.GetOrder(),
		"total_paid":      &totalPaid,
		"recommendations": recommendations,
		"loggedin":        loggedin,
		"username":        user,
	}); err != nil {
		log.Println(err)
	}
}

func (fe *frontendServer) logoutHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.Debug("logging out")
	w.Header().Set("Location", "/")

	for _, c := range r.Cookies() {
		c.Expires = time.Now().Add(-time.Hour * 24 * 365)
		c.MaxAge = -1
		http.SetCookie(w, c)
	}
	resp, err := http.Get("http://" + fe.userDbSvcAddr + "/logout")

	if err != nil {
		log.Fatal(err)
		renderHTTPError(log, r, w, errors.Wrap(err, "Error logging out!"), http.StatusBadRequest)
	}
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	if res["status_code"] != nil && int(res["status_code"].(float64)) == 200 {

		loggedin = false
		user = ""
	}

	http.Redirect(w, r, "/user", http.StatusFound)
	w.WriteHeader(http.StatusFound)
}

func (fe *frontendServer) setCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	cur := r.FormValue("currency_code")
	log.WithField("curr.new", cur).WithField("curr.old", currentCurrency(r)).
		Debug("setting currency")

	if cur != "" {
		http.SetCookie(w, &http.Cookie{
			Name:   cookieCurrency,
			Value:  cur,
			MaxAge: cookieMaxAge,
		})
	}
	referer := r.Header.Get("referer")
	if referer == "" {
		referer = "/"
	}
	w.Header().Set("Location", referer)
	w.WriteHeader(http.StatusFound)
}

// chooseAd queries for advertisements available and randomly chooses one, if
// available. It ignores the error retrieving the ad since it is not critical.
func (fe *frontendServer) chooseAd(ctx context.Context, ctxKeys []string, log logrus.FieldLogger) *pb.Ad {
	ads, err := fe.getAd(ctx, ctxKeys)
	if err != nil {
		log.WithField("error", err).Warn("failed to retrieve ads")
		return nil
	}
	return ads[rand.Intn(len(ads))]
}

func renderHTTPError(log logrus.FieldLogger, r *http.Request, w http.ResponseWriter, err error, code int) {
	log.WithField("error", err).Error("request error")
	errMsg := fmt.Sprintf("%s  %+v\n%s", "Error in the server(34.141.200.172): ", err, "go v1.11.5 linux/amd64")

	w.WriteHeader(code)
	if templateErr := templates.ExecuteTemplate(w, "error", map[string]interface{}{
		"session_id":  sessionID(r),
		"request_id":  r.Context().Value(ctxKeyRequestID{}),
		"error":       errMsg,
		"status_code": code,
		"status":      http.StatusText(code),
		"loggedin":    loggedin,
		"username":    user,
	}); templateErr != nil {
		log.Println(templateErr)
	}
}

func (fe *frontendServer) userHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.Debug("User Account Login/ Signup")
	w.Header().Set("location", "/user")

	cart, err := fe.getCart(r.Context(), sessionID(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve cart"), http.StatusInternalServerError)
		return
	}

	credit_card := "---"

	if loggedin {
		data := map[string]string{
			"username": user,
		}
		json_data, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := http.Post("http://"+fe.userDbSvcAddr+"/get_user_info", "application/json", bytes.NewBuffer(json_data))

		if err != nil {
			log.Fatal(err)
			renderHTTPError(log, r, w, errors.Wrap(err, "Error getting user info!"), http.StatusBadRequest)
		}

		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Println(res["credit_card"])
		if res["credit_card"] != nil {
			credit_card = res["credit_card"].(string)
			cc_split := strings.Split(credit_card, "-")
			for i := 0; i < 3; i++ {
				credit_card = strings.Replace(credit_card, cc_split[i]+"-", "****", 1)
			}
		}
	}

	if err := templates.ExecuteTemplate(w, "user", map[string]interface{}{
		"session_id":    sessionID(r),
		"request_id":    r.Context().Value(ctxKeyRequestID{}),
		"user_currency": currentCurrency(r),
		"show_currency": false,
		"cart_size":     cartSize(cart),
		"loggedin":      loggedin,
		"username":      user,
		"password":      "********",
		"credit_card":   credit_card,
	}); err != nil {
		log.Println(err)
	}
	w.WriteHeader(http.StatusFound)
}

func (fe *frontendServer) userLoginHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	// fmt.Println("method:", r.Method) //get request method
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("login_user")
		password := r.Form.Get("login_pass")
		fmt.Println("username:", username)
		fmt.Println("password:", password)

		data := map[string]string{
			"username": username,
			"password": password,
		}
		json_data, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := http.Post("http://"+fe.userDbSvcAddr+"/login", "application/json", bytes.NewBuffer(json_data))

		if err != nil {
			log.Fatal(err)
			renderHTTPError(log, r, w, errors.Wrap(err, "Error logging in!"), http.StatusBadRequest)
		}
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		fmt.Println(res["status_code"])
		if res["status_code"] != nil && int(res["status_code"].(float64)) == 200 {
			loggedin = true
			user = username
			// fmt.Println(user)
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			if res["status_code"] != nil && res["message"] != nil {
				renderHTTPError(log, r, w, errors.New(res["message"].(string)), int(res["status_code"].(float64)))
			}
		}

	}
}

func (fe *frontendServer) userSignUpHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	// fmt.Println("method:", r.Method) //get request method
	if r.Method == "POST" {
		r.ParseForm()
		username := strings.TrimSpace(r.Form.Get("signup_user"))
		password := strings.TrimSpace(r.Form.Get("signup_pass"))
		repeat_password := strings.TrimSpace(r.Form.Get("signup_repeat_pass"))
		if username == "" || password == "" || repeat_password == "" {
			log.Warn("One of the fields is empty!")
			http.Redirect(w, r, "/user", http.StatusFound)
			return
		}

		if password != repeat_password {
			log.Warn("Passwords do not match!")
			renderHTTPError(log, r, w, errors.New("The two password entries do not match!"), http.StatusBadRequest)
			return
		}

		fmt.Println("username:", username)
		fmt.Println("password:", password)

		data := map[string]string{
			"username": username,
			"password": password,
		}
		json_data, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := http.Post("http://"+fe.userDbSvcAddr+"/register", "application/json", bytes.NewBuffer(json_data))
		if err != nil {
			log.Fatal(err)
			renderHTTPError(log, r, w, errors.Wrap(err, "Error during Sign Up!"), http.StatusBadRequest)
		}
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)

		if res["status_code"] != nil && int(res["status_code"].(float64)) == 200 {
			loggedin = true
			user = username
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			if res["status_code"] != nil && res["message"] != nil {
				renderHTTPError(log, r, w, errors.New(res["message"].(string)), int(res["status_code"].(float64)))
			}
		}
	}
}

func currentCurrency(r *http.Request) string {
	c, _ := r.Cookie(cookieCurrency)
	if c != nil {
		return c.Value
	}
	return defaultCurrency
}

func sessionID(r *http.Request) string {
	v := r.Context().Value(ctxKeySessionID{})
	if v != nil {
		return v.(string)
	}
	return ""
}

func cartIDs(c []*pb.CartItem) []string {
	out := make([]string, len(c))
	for i, v := range c {
		out[i] = v.GetProductId()
	}
	return out
}

// get total # of items in cart
func cartSize(c []*pb.CartItem) int {
	cartSize := 0
	for _, item := range c {
		cartSize += int(item.GetQuantity())
	}
	return cartSize
}

func renderMoney(money pb.Money) string {
	return fmt.Sprintf("%s %d.%02d", money.GetCurrencyCode(), money.GetUnits(), money.GetNanos()/10000000)
}

func randNum() string {
	randstr := ""
	for i := 0; i < 4; i++ {
		rand_val := rand.Intn(10)
		randstr += strconv.Itoa(rand_val)
	}
	return randstr
}
