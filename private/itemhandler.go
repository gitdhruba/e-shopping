package private

import (
	//"math/rand"

	"fmt"

	db "main.go/database"
	"main.go/models"

	//"main.go/util"

	//"golang.org/x/crypto/bcrypt"

	//"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

//function for entering purchased item details in DB
func CreateEntry(c *fiber.Ctx) error {

	type iteminput struct {
		Bookid   uint32 `json:"bookid"`
		Quantity uint32 `json:"quantity"`
	}

	input := new(iteminput)
	if err := c.BodyParser(input); err != nil {
		return c.JSON(fiber.Map{
			"error": true,
			"msg":   "incorrect input",
		})
	}
	//fmt.Println(models.VerifiedUser)
	var book models.BookStock
	res := db.DB.Where("bookid = ?", input.Bookid).Find(&book)
	res.Scan(&book)

	t := models.GetTimeNow()
	fmt.Println(t)
	item := models.Item{
		User:       fmt.Sprint(models.VerifiedUser),
		Bookid:     book.Bookid,
		Bookname:   book.Bookname,
		Time:       t,
		Quantity:   input.Quantity,
		Totalprice: (uint64(input.Quantity) * (book.Price)),
	}

	fmt.Println(book)
	fmt.Println(item)

	if book.Quantity < input.Quantity {
		return c.JSON(fiber.Map{
			"error": true,
			"msg":   "out of stock",
		})
	} else {
		book.Quantity = book.Quantity - input.Quantity
		err := db.DB.Model(&models.BookStock{}).Where("bookid = ?", book.Bookid).Update("quantity", book.Quantity).Error
		if err != nil {
			return c.JSON(fiber.Map{
				"error": true,
				"msg":   "update error",
			})
		}
	}

	if err := db.DB.Create(&item).Error; err != nil {
		return c.JSON(fiber.Map{
			"error": true,
			"msg":   "insertion error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":  false,
		"status": "order successfull",
	})
}

//function for removing purchased item data
func DeleteEntry(c *fiber.Ctx) error {

	type iteminput struct {
		Bookid uint32 `json:"bookid"`
		Time   string `json:"time"`
	}

	input := new(iteminput)
	if err := c.BodyParser(input); err != nil {
		return c.JSON(fiber.Map{
			"error":  true,
			"status": "incorrect input",
		})
	}

	var book models.BookStock
	resbook := db.DB.Where("bookid = ?", input.Bookid).Find(&models.BookStock{})
	resbook.Scan(&book)
	fmt.Println(book)

	var item models.Item
	resitem := db.DB.Where("\"user\" = ? AND bookid = ? AND \"time\" = ?", models.VerifiedUser, book.Bookid, input.Time).Find(&item)
	resitem.Scan(&item)
	fmt.Println(item)

	book.Quantity = book.Quantity + item.Quantity
	err := db.DB.Model(&models.BookStock{}).Where("bookid = ?", input.Bookid).Update("quantity", book.Quantity).Error
	if err != nil {
		return c.JSON(fiber.Map{
			"error":  true,
			"status": "update error",
		})
	}

	if err := db.DB.Where("\"user\" = ? AND bookid = ? AND \"time\" = ?", models.VerifiedUser, book.Bookid, input.Time).Delete(&models.Item{}).Error; err != nil {
		return c.JSON(fiber.Map{
			"error":  true,
			"status": "Deletion error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":  false,
		"status": "cancellation successfull",
	})
}

//function for adding items to the cart
func AddtoCart(c *fiber.Ctx) error {

	type iteminput struct {
		Bookid   uint32 `json:"bookid"`
		Quantity uint32 `json:"quantity"`
	}

	input := new(iteminput)
	if err := c.BodyParser(input); err != nil {
		return c.JSON(fiber.Map{
			"error": true,
			"msg":   "incorrect input",
		})
	}

	var book models.BookStock
	var cart models.Cart
	res := db.DB.Where("bookid = ?", input.Bookid).Find(&models.BookStock{})
	res.Scan(&book)

	t := models.GetTimeNow()
	fmt.Println(t)
	if r := db.DB.Where("\"user\" = ? AND bookid = ?", models.VerifiedUser, book.Bookid).Find(&models.Cart{}); r.RowsAffected <= 0 {
		cartitem := models.Cart{
			User:       fmt.Sprint(models.VerifiedUser),
			Bookid:     input.Bookid,
			Bookname:   book.Bookname,
			Time:       t,
			Quantity:   input.Quantity,
			Totalprice: (uint64(input.Quantity * uint32(book.Price))),
		}

		if book.Quantity < input.Quantity {
			return c.JSON(fiber.Map{
				"error": true,
				"msg":   "not having enough stock",
			})
		}

		if err := db.DB.Create(&cartitem).Error; err != nil {
			return c.JSON(fiber.Map{
				"error": true,
				"msg":   "insertion error",
			})
		}

	} else {
		r.Scan(&cart)
		cart.Quantity = cart.Quantity + input.Quantity
		if book.Quantity < cart.Quantity {
			return c.JSON(fiber.Map{
				"error": true,
				"msg":   "not having enough stock",
			})
		}
		cart.Totalprice = uint64(cart.Quantity) * book.Price
		if err := db.DB.Model(&models.Cart{}).Where("\"user\" = ? AND bookid = ?", models.VerifiedUser, book.Bookid).Updates(models.Cart{Time: t, Quantity: cart.Quantity, Totalprice: cart.Totalprice}).Error; err != nil {
			return c.JSON(fiber.Map{
				"error": true,
				"msg":   "insertion error",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"msg":   "adding to cart successfull",
	})

}

//function for removing items from cart
func DeletefromCart(c *fiber.Ctx) error {

	type iteminput struct {
		Bookid uint32 `json:"bookid"`
		Time   string `json:"time"`
	}

	input := new(iteminput)
	if err := c.BodyParser(input); err != nil {
		return c.JSON(fiber.Map{
			"error":  true,
			"status": "incorrect input",
		})
	}

	err := db.DB.Where("\"user\" = ? AND bookid = ? AND \"time\" = ?", models.VerifiedUser, input.Bookid, input.Time).Delete(&models.Cart{}).Error
	if err != nil {
		fmt.Println("error occured while deleting cart")
		return c.JSON(fiber.Map{
			"error": true,
			"msg":   "deletion error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":  false,
		"status": "cart deletion successfull",
	})
}

//function to place cart orders
func PlaceCartOrders(c *fiber.Ctx) error {

	var book models.BookStock
	var cart models.Cart
	var skippeditems []models.Cart
	var qnt uint32

	rows, _ := db.DB.Model(&models.Cart{}).Where("\"user\" = ?", models.VerifiedUser).Rows()
	defer rows.Close()

	for rows.Next() {

		db.DB.ScanRows(rows, &cart)
		res := db.DB.Select("quantity").Where("bookid = ?", cart.Bookid).Find(&book)
		res.Scan(&qnt)

		if qnt < cart.Quantity {
			skippeditems = append(skippeditems, cart)
			continue
		}

		qnt = qnt - cart.Quantity
		t := models.GetTimeNow()

		item := models.Item{
			User:       cart.User,
			Bookid:     cart.Bookid,
			Bookname:   cart.Bookname,
			Time:       t,
			Quantity:   cart.Quantity,
			Totalprice: cart.Totalprice,
		}

		if err := db.DB.Model(&book).Where("bookid = ?", cart.Bookid).Update("quantity", qnt).Error; err != nil {
			fmt.Println("update error")
		}

		if err := db.DB.Create(&item).Error; err != nil {
			fmt.Println("insert error")
		}

	}

	if err := db.DB.Where("\"user\" = ?", models.VerifiedUser).Delete(&models.Cart{}).Error; err != nil {
		fmt.Println("deletion error")
	}

	fmt.Println(skippeditems)

	return c.Status(fiber.StatusOK).JSON(skippeditems)
}

//function for getting cart data
func GetCartData(c *fiber.Ctx) error {
	type cartdata struct {
		Bookid     uint32
		Bookname   string
		Time       string
		Quantity   uint32
		Totalprice uint64
	}

	var cartitems []cartdata
	var cartitem cartdata
	var cart models.Cart

	rows, _ := db.DB.Model(&models.Cart{}).Where("\"user\" = ?", models.VerifiedUser).Rows()
	defer rows.Close()
	for rows.Next() {
		db.DB.ScanRows(rows, &cart)
		cartitem.Bookid = cart.Bookid
		cartitem.Bookname = cart.Bookname
		cartitem.Time = cart.Time
		cartitem.Quantity = cart.Quantity
		cartitem.Totalprice = cart.Totalprice
		cartitems = append(cartitems, cartitem)
	}
	fmt.Println(cartitems)

	return c.JSON(cartitems)
}

//function for getting purchased items data
func GetPurchaseData(c *fiber.Ctx) error {
	type purchasedata struct {
		Bookid     uint32
		Bookname   string
		Time       string
		Quantity   uint32
		Totalprice uint64
	}

	var purchaseditems []purchasedata
	var purchaseditem purchasedata
	var item models.Item

	rows, _ := db.DB.Model(&models.Item{}).Where("\"user\" = ?", models.VerifiedUser).Rows()
	defer rows.Close()
	for rows.Next() {
		db.DB.ScanRows(rows, &item)
		purchaseditem.Bookid = item.Bookid
		purchaseditem.Bookname = item.Bookname
		purchaseditem.Time = item.Time
		purchaseditem.Quantity = item.Quantity
		purchaseditem.Totalprice = item.Totalprice
		purchaseditems = append(purchaseditems, purchaseditem)
	}
	fmt.Println(purchaseditems)

	return c.JSON(purchaseditems)
}
