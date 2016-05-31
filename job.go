package cfbackup

func (s *Jobs) GetInstances() (i []Instances) {
	i = s.Instances

	if len(i) == 0 {
		i = append([]Instances{}, s.Instance)
	}
	return
}
