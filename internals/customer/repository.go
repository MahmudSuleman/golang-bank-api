package customer

type Repository interface {
	GetAll() []Customer
	GetById(id int) (Customer, bool)
	Create(customer Customer) Customer
}

type MemoryRepository struct {
	customers []Customer
	nextId    int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		customers: []Customer{
			{
				ID:        1,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			{
				ID:        2,
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane@example.com",
			},
		},
		nextId: 3,
	}
}

func (r *MemoryRepository) GetAll() []Customer {
	return r.customers
}

func (r *MemoryRepository) GetById(id int) (Customer, bool) {
	for _, customer := range r.customers {
		if customer.ID == id {
			return customer, true
		}
	}
	return Customer{}, false
}

func (r *MemoryRepository) Create(customer Customer) Customer {
	customer.ID = r.nextId
	r.customers = append(r.customers, customer)
	r.nextId++
	return customer
}
