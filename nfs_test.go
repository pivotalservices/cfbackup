package cfbackup_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	. "github.com/pivotalservices/cfbackup"

	"github.com/pivotalservices/cfbackup/fakes"
	"github.com/pivotalservices/gtils/command"
	"github.com/pivotalservices/gtils/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("nfs", func() {
	Describe("Import NFS Restore", func() {
		var (
			nfs           *NFSBackup
			buffer        *gbytes.Buffer
			controlString string = "test of local file"
			err           error
		)

		BeforeEach(func() {
			err = nil
		})

		AfterEach(func() {
			err = nil
		})

		Context("successful call to import", func() {

			BeforeEach(func() {
				lf := strings.NewReader(controlString)
				buffer = gbytes.NewBuffer()
				nfs = fakes.GetNfs(buffer, &fakes.SuccessMockNFSExecuter{})
				err = nfs.Import(lf)
			})

			It("should return nil error", func() {
				Ω(err).Should(BeNil())
			})

			It("should write the local file contents to the remote", func() {
				Ω(buffer).Should(gbytes.Say(controlString))
			})
		})

		Context("error on command execution", func() {

			BeforeEach(func() {
				lf := strings.NewReader(controlString)
				buffer = gbytes.NewBuffer()
				nfs = fakes.GetNfs(buffer, &fakes.FailureMockNFSExecuter{})
				err = nfs.Import(lf)
			})

			It("should return non-nil execution error", func() {
				Ω(err).ShouldNot(BeNil())
				Ω(err).Should(Equal(fakes.ErrMockNfsCommand))
			})

			It("should write the local file contents to the remote", func() {
				Ω(buffer).Should(gbytes.Say(controlString))
			})
		})

		Context("error on file upload", func() {
			BeforeEach(func() {
				buffer = gbytes.NewBuffer()
				nfs = fakes.GetNfs(buffer, &fakes.SuccessMockNFSExecuter{})
			})

			Context("Read failure", func() {
				BeforeEach(func() {
					lf := mock.NewReadWriteCloser(mock.ErrReadFailure, nil, nil)
					err = nfs.Import(lf)
				})

				It("should return non-nil execution error", func() {
					Ω(err).ShouldNot(BeNil())
					Ω(err).Should(Equal(mock.ErrReadFailure))
				})

				It("should write the local file contents to the remote", func() {
					Ω(buffer).ShouldNot(gbytes.Say(controlString))
				})
			})

			Context("Writer related failure", func() {
				Context("Write failure", func() {
					BeforeEach(func() {
						lf := mock.NewReadWriteCloser(nil, mock.ErrWriteFailure, nil)
						nfs = fakes.GetNfs(lf, &fakes.SuccessMockNFSExecuter{})
						err = nfs.Import(lf)
					})

					It("should return non-nil execution error", func() {
						Ω(err).ShouldNot(BeNil())
						Ω(err).Should(Equal(mock.ErrWriteFailure))
					})

					It("should write the local file contents to the remote", func() {
						Ω(buffer).ShouldNot(gbytes.Say(controlString))
					})
				})

				Context("Close failure", func() {
					BeforeEach(func() {
						lf := mock.NewReadWriteCloser(nil, nil, mock.ErrCloseFailure)
						nfs = fakes.GetNfs(lf, &fakes.SuccessMockNFSExecuter{})
						err = nfs.Import(lf)
					})

					It("should return non-nil execution error", func() {
						Ω(err).ShouldNot(BeNil())
						Ω(err).Should(Equal(io.ErrShortWrite))
					})

					It("should write the local file contents to the remote", func() {
						Ω(buffer).ShouldNot(gbytes.Say(controlString))
					})
				})
			})
		})
	})

	Describe("NFSBackup", func() {
		var nfs *NFSBackup

		BeforeEach(func() {
			nfs = &NFSBackup{}
		})

		Context("sucessfully calling Dump", func() {
			BeforeEach(func() {
				nfs.Caller = &fakes.SuccessMockNFSExecuter{}
			})

			It("Should return nil error and a success message in the writer", func() {
				var b bytes.Buffer
				err := nfs.Dump(&b)
				Ω(err).Should(BeNil())
				Ω(b.String()).Should(Equal(fakes.NfsSuccessString))
			})
		})

		Context("failed calling Dump", func() {
			BeforeEach(func() {
				nfs.Caller = &fakes.FailureMockNFSExecuter{}
			})

			It("Should return non nil error and a failure output in the writer", func() {
				var b bytes.Buffer
				err := nfs.Dump(&b)
				Ω(err).ShouldNot(BeNil())
				Ω(b.String()).Should(Equal(fakes.NfsFailureString))
			})
		})

		Describe("NewNFSBackup", func() {
			Context("when executer is created successfully", func() {
				var origExecuterFunction func(command.SshConfig) (command.Executer, error)

				BeforeEach(func() {
					origExecuterFunction = NfsNewRemoteExecuter
					NfsNewRemoteExecuter = func(command.SshConfig) (command.Executer, error) {
						return &fakes.SuccessMockNFSExecuter{}, nil
					}
				})

				AfterEach(func() {
					NfsNewRemoteExecuter = origExecuterFunction
				})

				It("should return a nil error and a non-nil NFSBackup object", func() {
					n, err := NewNFSBackup("pass", "0.0.0.0", "", "/var/somepath")
					Ω(err).Should(BeNil())
					Ω(n).Should(BeAssignableToTypeOf(&NFSBackup{}))
					Ω(n).ShouldNot(BeNil())
				})
			})

			Context("when executer fails to be created properly", func() {
				var origExecuterFunction func(command.SshConfig) (command.Executer, error)

				BeforeEach(func() {
					origExecuterFunction = NfsNewRemoteExecuter
					NfsNewRemoteExecuter = func(command.SshConfig) (ce command.Executer, err error) {
						ce = &fakes.FailureMockNFSExecuter{}
						err = fmt.Errorf("we have an error")
						return
					}
				})

				AfterEach(func() {
					NfsNewRemoteExecuter = origExecuterFunction
				})

				It("should return a nil error and a NFSBackup object that is nil", func() {
					n, err := NewNFSBackup("pass", "0.0.0.0", "", "/var/somepath")
					Ω(err).ShouldNot(BeNil())
					Ω(n).Should(BeNil())
					Ω(n).Should(BeAssignableToTypeOf(&NFSBackup{}))
					Ω(n).Should(BeNil())
				})
			})
		})
	})
})
