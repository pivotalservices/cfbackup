package cfbackup

import "fmt"

//FindIPsByProductGUIDAndJobGUIDAndAvailabilityZoneGUID - returns array of IPs based on product and job guids
func (s *IPAssignments) FindIPsByProductGUIDAndJobGUIDAndAvailabilityZoneGUID(productGUID, jobGUID string, azGUID string) (ips []string, err error) {
	if s.containsProductGUID(productGUID) {
        productMap := s.Assignments[productGUID]
        if s.containsJobGUID(productMap, jobGUID) {
            jobMap := productMap[jobGUID]  
            if s.containsAZGUID(jobMap, azGUID) {
                ips = jobMap[azGUID]
            } else {
                err = fmt.Errorf("AZ guid not found %s", azGUID)
            }
            
        } else {
            err = fmt.Errorf("Job guid not found %s", jobGUID)
        }
        
	} else {
		err = fmt.Errorf("Product guid not found %s", productGUID)
	}
	return
}

func (s *IPAssignments) containsProductGUID(productGUID string) bool {
	_, ok := s.Assignments[productGUID]
	return ok
}

func (s *IPAssignments) containsJobGUID(productMap map[string]map[string][]string, jobGUID string) (bool) {
	_, ok := productMap[jobGUID]
	return ok
}

func (s *IPAssignments) containsAZGUID(jobMap map[string][]string, azGUID string) (bool) {
	_, ok := jobMap[azGUID]
	return ok
}

