role :app, "am122" # sandbox
role :db,  "am122", :primary => true # This is where Rails migrations will run
role :web, "am122"                          # Your HTTP server, Apache/etc
set :user, 'qboxserver'
set :repository,  "http://ci.qbox.me/jenkins/job/#{application}-production"
set :deploy_to do "/home/#{user}/websites/mmq" end
set :jenkins_artifact_file, 'dist/qboxmmq'
set :conf_name, 'qboxmmq'


namespace :deploy do
  task :restart, :roles => :app, :max_hosts => 1 do
    run 'supervisorctl restart mmq'
  end
end


