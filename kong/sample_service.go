package kong

type SampleService service

func (s *SampleService) Foo() {
	s.client.Do()
}
