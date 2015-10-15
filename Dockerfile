FROM centurylink/ca-certs

ADD AlertAPI /AlertAPI
ENTRYPOINT [ "/AlertAPI" ]
EXPOSE 8080
