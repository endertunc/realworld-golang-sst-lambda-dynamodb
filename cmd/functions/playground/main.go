package main

type Address struct {
	Street string
	City   string
	State  string
	Zip    string
}

func main() {

	//
	//err := errutil.ErrEmailAlreadyExists.Errorf("can't find the email man: %w", errors.New("dynamodb goes brrrr"))
	//
	//appErr := errutil.AppError{}
	//okAs := errors.As(err, &appErr)
	//
	//okIs := errors.Is(appErr, errutil.ErrEmailAlreadyExists)
	//
	//if okAs {
	//	fmt.Println("okAs")
	//} else {
	//	fmt.Println("not okAs")
	//}
	//
	//if okIs {
	//	fmt.Println("okIs")
	//} else {
	//	fmt.Println("not okIs")
	//}

	//old()

}

//func old() {
//	rootErr := errors.New("fuck you asshole")
//
//	err := security.ErrTokenSubjectInvalid.Errorf("%w", rootErr)
//	app := errutil.AppError{}
//	result1 := errors.As(err, &app)
//
//	result2 := errors.Is(app, security.ErrTokenSubjectInvalid)
//
//	fmt.Println(result1)
//	fmt.Println(result2)
//
//	fmt.Println(app.Error())
//	fmt.Println(app.Message)
//	fmt.Println(app.InternalMessage)
//	fmt.Println(app.Cause.Error())
//}
