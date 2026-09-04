package customer

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetAll() []Customer {
	return s.repository.GetAll()
}

func (s *Service) GetById(id int) (Customer, bool) {
	return s.repository.GetById(id)
}

func (s *Service) Create(customer Customer) Customer {
	return s.repository.Create(customer)
}
